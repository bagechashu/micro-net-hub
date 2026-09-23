package bot

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

// stopTimeout 单个实例停止的等待上限。
//
// Provider 的 Start 在 ctx 取消后仍可能阻塞在一次在途的长轮询上，超时后放弃
// 等待并继续处理其余实例，避免一个卡住的实例拖住整个热加载请求。
const stopTimeout = 10 * time.Second

// InstanceSpec 一个 Bot 实例的运行规格。
//
// 独立于 config.BotInstanceConfig 定义，使本包不反向依赖配置包；由调用方
// （bootstrap / web）负责配置结构到运行规格的转换。
type InstanceSpec struct {
	ID     string
	Type   string
	Name   string
	Config map[string]string
}

// fingerprint 计算实例规格的指纹，用于热加载时判断该实例是否需要重建。
//
// 只有指纹变化的实例才会被停掉重建，未改动的实例不受影响——否则每次保存配置
// 都会重启全部机器人，正在处理中的会话会被打断。
func (s InstanceSpec) fingerprint() string {
	keys := make([]string, 0, len(s.Config))
	for k := range s.Config {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString(s.Type)
	b.WriteByte('\n')
	b.WriteString(s.Name)
	for _, k := range keys {
		b.WriteByte('\n')
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(s.Config[k])
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

// Status 一个运行中实例的对外状态快照
type Status struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// instance 一个运行中的实例及其生命周期控制
type instance struct {
	provider    BotProvider
	cancel      context.CancelFunc
	done        chan struct{}
	fingerprint string
}

// Manager 管理一组运行中的 Bot 实例，并支持按配置差分热加载。
//
// 每个实例持有独立的子 context，因此可以单独停止/重建而不影响其他实例；这是
// "面板改完 Bot 配置立即生效"与"单实例失败不拖垮其他实例"两个要求的实现基础。
type Manager struct {
	registry *Registry
	handler  MessageHandler

	mu      sync.Mutex
	baseCtx context.Context
	running map[string]*instance
}

// MessageHandler 处理一条来自某个实例的消息。
//
// 显式传入 provider 而非让上层去查表：回复必须发回消息来源实例，把它作为参数
// 交出去，多实例串台在类型层面就无从发生。ctx 是该实例自己的子 context，实例被
// 热移除时在途处理会一并取消。
type MessageHandler func(ctx context.Context, bot BotProvider, msg Message)

// NewManager 创建实例管理器。
//
// registry 与 handler 均为必需依赖，缺失属装配错误，直接 panic 而非留到运行期
// 静默丢消息。
func NewManager(registry *Registry, handler MessageHandler) *Manager {
	if registry == nil {
		panic("bot: NewManager 的 registry 不能为 nil")
	}
	if handler == nil {
		panic("bot: NewManager 的 handler 不能为 nil")
	}
	return &Manager{
		registry: registry,
		handler:  handler,
		baseCtx:  context.Background(),
		running:  make(map[string]*instance),
	}
}

// Apply 将运行中的实例集合收敛到 specs 描述的目标状态。
//
// 差分策略：不在 specs 中的实例停止并移除；指纹变化的实例停止后重建；指纹未变
// 的实例原样保留。单个实例创建失败只记录该实例的错误并继续处理其余实例（与项目
// "单个工具失败不阻塞其他工具"一致），全部错误汇总返回，由调用方决定如何呈现。
//
// 首次调用前必须已 Start（baseCtx 就绪），否则新实例会绑定到 Background 上而
// 无法随进程退出停止。
func (m *Manager) Apply(specs []InstanceSpec) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	desired := make(map[string]InstanceSpec, len(specs))
	for _, spec := range specs {
		desired[spec.ID] = spec
	}

	// 1. 停掉已删除或配置已变化的实例
	for id, inst := range m.running {
		spec, keep := desired[id]
		if keep && spec.fingerprint() == inst.fingerprint {
			continue
		}
		reason := "配置已变更，重建"
		if !keep {
			reason = "已从配置移除，停止"
		}
		log.Printf("[Bot] 实例 [%s] %s", id, reason)
		m.stopLocked(inst)
		delete(m.running, id)
	}

	// 2. 创建尚未运行的实例
	var errs []error
	for _, spec := range specs {
		if _, ok := m.running[spec.ID]; ok {
			continue
		}
		if err := m.startLocked(spec); err != nil {
			errs = append(errs, err)
			log.Printf("[Bot] ⚠️ 实例 [%s] 启动失败: %v", spec.ID, err)
			continue
		}
		log.Printf("[Bot] ✓ 实例 [%s] (%s) 已启动", spec.ID, spec.Type)
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d/%d 个 Bot 实例启动失败: %w", len(errs), len(specs), errors.Join(errs...))
	}
	return nil
}

// startLocked 创建并启动单个实例，调用方必须持有 m.mu
func (m *Manager) startLocked(spec InstanceSpec) error {
	provider, err := m.registry.Create(spec.Type, spec.ID, spec.Name, spec.Config)
	if err != nil {
		return fmt.Errorf("bot 实例 [%s] 初始化失败: %w", spec.ID, err)
	}

	ctx, cancel := context.WithCancel(m.baseCtx)
	// 处理器在 Start 之前注册，保证实例开始收消息时处理链路已就位
	provider.RegisterHandler(func(msg Message) {
		m.handler(ctx, provider, msg)
	})

	inst := &instance{
		provider:    provider,
		cancel:      cancel,
		done:        make(chan struct{}),
		fingerprint: spec.fingerprint(),
	}

	go func() {
		defer close(inst.done)
		if err := provider.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("[Bot] 实例 [%s] 监听退出: %v", spec.ID, err)
		}
	}()

	m.running[spec.ID] = inst
	return nil
}

// stopLocked 停止单个实例并等待其监听 goroutine 退出，调用方必须持有 m.mu
func (m *Manager) stopLocked(inst *instance) {
	inst.cancel()
	select {
	case <-inst.done:
	case <-time.After(stopTimeout):
		log.Printf("[Bot] ⚠️ 实例 [%s] 停止超时 (%s)，放弃等待", inst.provider.ID(), stopTimeout)
	}
}

// Start 绑定进程级 context 并按 specs 启动首批实例。
//
// baseCtx 派生出的子 context 让实例既能被 Apply 单独停止，也能随进程退出统一停止。
func (m *Manager) Start(ctx context.Context, specs []InstanceSpec) error {
	m.mu.Lock()
	m.baseCtx = ctx
	m.mu.Unlock()
	return m.Apply(specs)
}

// StopAll 停止全部实例，供进程优雅退出时调用
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, inst := range m.running {
		m.stopLocked(inst)
		delete(m.running, id)
	}
}

// Statuses 返回当前运行中实例的状态快照，按实例 ID 排序
func (m *Manager) Statuses() []Status {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]Status, 0, len(m.running))
	for _, inst := range m.running {
		out = append(out, Status{
			ID:   inst.provider.ID(),
			Type: inst.provider.Type(),
			Name: inst.provider.Name(),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Providers 返回全部运行中实例的 Provider, 按实例 ID 排序.
//
// 供需要向"所有实例"广播的调用方使用(如审批通知在未限定 instance-ids 时),
// 语义上与按 ID 解析的 Provider 互补.
func (m *Manager) Providers() []BotProvider {
	m.mu.Lock()
	defer m.mu.Unlock()

	ids := make([]string, 0, len(m.running))
	for id := range m.running {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	out := make([]BotProvider, 0, len(ids))
	for _, id := range ids {
		out = append(out, m.running[id].provider)
	}
	return out
}
