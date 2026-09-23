// Package bot 提供统一的即时通讯接入抽象：BotProvider 接口、Bot 类型注册表与
// 多实例运行时管理器。
//
// 新增一个 Bot 平台只需两步：实现 BotProvider 接口，并在本包内用 Register 把
// TypeSpec + 工厂函数登记进注册表。上层（bootstrap / web）只依赖接口与注册表，
// 无需为新平台改动任何编排、审计或配置代码。
//
// 类型专属知识（必填配置项、敏感配置项）随 TypeSpec 收敛在本包，避免"新增类型
// 要同时改 validator、前端表单、脱敏名单"这种散落。
package bot

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Message 定义统一的消息结构体
type Message struct {
	BotID     string // 来源 Bot 实例 ID，用于审计归档与限速分桶
	BotType   string // 来源 Bot 类型（telegram / feishu / wecom ...）
	UserID    string // 统一映射为字符串类型的用户 ID
	Username  string // 用户昵称
	ChannelID string // 聊天渠道/群组 ID
	ChatType  string // 聊天类型（private / group / supergroup ...），审批等安全边界的判定依据
	Content   string // 消息纯文本内容
}

// BotProvider 统一的即时通讯对接接口。
//
// Start 必须阻塞直到 ctx 被取消，并在返回前释放自身的长连接/长轮询资源；
// Manager 依赖这一约定实现单个实例的热重启与热移除。
type BotProvider interface {
	ID() string                                                          // 实例唯一标识
	Type() string                                                        // Bot 类型
	Name() string                                                        // 显示名称
	Start(ctx context.Context) error                                     // 启动监听，阻塞至 ctx 取消
	SendMessage(ctx context.Context, toID string, text string) error     // 发送普通文本消息
	SendHTMLMessage(ctx context.Context, toID string, html string) error // 发送带格式的 HTML 消息
	SendTyping(ctx context.Context, toID string) error                   // 发送"正在输入"状态（瞬时，需按平台节奏重发）
	RegisterHandler(handler func(msg Message))                           // 注册消息处理器
}

// TypingInterval "正在输入"状态的建议重发间隔。
//
// Telegram 的 ChatAction 约 5 秒后自动失效，取其安全余量 4 秒；调用方在长耗时请求
// 处理期间应按此间隔反复调用 SendTyping，保持状态常新。
const TypingInterval = 4 * time.Second

// Factory 按实例配置创建 BotProvider 的工厂函数
type Factory func(id, name string, config map[string]string) (BotProvider, error)

// TypeSpec 描述一种 Bot 类型的配置规格。
//
// RequiredKeys 供启动校验与写配置前校验共用；SensitiveKeys 供 Web API 脱敏与
// "占位值不覆盖原值"的合并逻辑共用；两者一并由前端用于渲染实例表单。
type TypeSpec struct {
	Type          string   `json:"type"`           // 类型标识，如 telegram
	Label         string   `json:"label"`          // 展示名，如 Telegram
	RequiredKeys  []string `json:"required_keys"`  // 必填的 config key
	OptionalKeys  []string `json:"optional_keys"`  // 可选的 config key
	SensitiveKeys []string `json:"sensitive_keys"` // 需脱敏的 config key
}

// IsSensitive 判断某个 config key 是否为该类型的敏感项
func (s TypeSpec) IsSensitive(key string) bool {
	for _, k := range s.SensitiveKeys {
		if k == key {
			return true
		}
	}
	return false
}

// registration 一种 Bot 类型的规格与工厂函数
type registration struct {
	spec    TypeSpec
	factory Factory
}

// Registry Bot 类型注册表，把类型标识映射到规格与工厂函数。
//
// 带锁：注册发生在包初始化阶段，但 Web 层的配置校验与表单元数据接口会在请求
// goroutine 中并发读取。
type Registry struct {
	mu    sync.RWMutex
	types map[string]registration
}

// NewRegistry 创建一个空的注册表（供测试独立装配；生产走 DefaultRegistry）
func NewRegistry() *Registry {
	return &Registry{types: make(map[string]registration)}
}

// Register 登记一种 Bot 类型的规格与工厂函数。
//
// spec.Type 为空或 factory 为 nil 属编程错误，直接 panic：注册发生在包初始化
// 阶段，静默忽略会让该类型在运行期表现为"配置看着合法但实例创建不出来"。
func (r *Registry) Register(spec TypeSpec, factory Factory) {
	if spec.Type == "" {
		panic("bot: Register 的 spec.Type 不能为空")
	}
	if factory == nil {
		panic("bot: Register 的 factory 不能为 nil, type=" + spec.Type)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.types[spec.Type] = registration{spec: spec, factory: factory}
}

// Create 按类型创建 BotProvider 实例，创建前先校验配置项完整性
func (r *Registry) Create(botType, id, name string, config map[string]string) (BotProvider, error) {
	r.mu.RLock()
	reg, ok := r.types[botType]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("未注册的 Bot 类型 %q（已支持: %v）", botType, r.SupportedTypes())
	}

	if err := validateAgainstSpec(reg.spec, config); err != nil {
		return nil, err
	}
	return reg.factory(id, name, config)
}

// ValidateConfig 校验某类型的配置项是否满足其 TypeSpec，供写配置前预检使用
func (r *Registry) ValidateConfig(botType string, config map[string]string) error {
	r.mu.RLock()
	reg, ok := r.types[botType]
	r.mu.RUnlock()
	if !ok {
		return fmt.Errorf("未注册的 Bot 类型 %q（已支持: %v）", botType, r.SupportedTypes())
	}
	return validateAgainstSpec(reg.spec, config)
}

// Spec 返回某类型的规格；类型未注册时 ok 为 false
func (r *Registry) Spec(botType string) (TypeSpec, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	reg, ok := r.types[botType]
	return reg.spec, ok
}

// Specs 返回全部已注册类型的规格，按类型标识排序（供前端渲染实例表单）
func (r *Registry) Specs() []TypeSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()

	specs := make([]TypeSpec, 0, len(r.types))
	for _, reg := range r.types {
		specs = append(specs, reg.spec)
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].Type < specs[j].Type })
	return specs
}

// SupportedTypes 返回全部已注册的类型标识，按字典序排序
func (r *Registry) SupportedTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.types))
	for t := range r.types {
		types = append(types, t)
	}
	sort.Strings(types)
	return types
}

// validateAgainstSpec 校验配置项是否覆盖了该类型的全部必填 key
func validateAgainstSpec(spec TypeSpec, config map[string]string) error {
	missing := make([]string, 0, len(spec.RequiredKeys))
	for _, key := range spec.RequiredKeys {
		if config[key] == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%s 实例缺少必填配置项: %v", spec.Type, missing)
	}
	return nil
}

// defaultRegistry 进程级注册表，内置类型在此登记，bootstrap 与 web 共用同一份
var defaultRegistry = func() *Registry {
	r := NewRegistry()
	r.Register(telegramSpec, newTelegramProvider)
	return r
}()

// DefaultRegistry 返回已内置全部自带 Bot 类型的进程级注册表
func DefaultRegistry() *Registry {
	return defaultRegistry
}
