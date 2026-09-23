package bot

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingRegistry 在 fake 类型之上记录每次创建出的 provider，便于断言实例被重建
type recordingRegistry struct {
	*Registry
	mu        sync.Mutex
	created   []*fakeProvider
	failOnIDs map[string]bool
}

func newRecordingRegistry(failOnIDs ...string) *recordingRegistry {
	rr := &recordingRegistry{Registry: NewRegistry(), failOnIDs: map[string]bool{}}
	for _, id := range failOnIDs {
		rr.failOnIDs[id] = true
	}

	rr.Register(fakeSpec, func(id, name string, config map[string]string) (BotProvider, error) {
		rr.mu.Lock()
		defer rr.mu.Unlock()
		if rr.failOnIDs[id] {
			return nil, fmt.Errorf("模拟实例 %s 认证失败", id)
		}
		p := &fakeProvider{id: id, botType: "fake", name: name, started: make(chan struct{})}
		rr.created = append(rr.created, p)
		return p, nil
	})
	return rr
}

func (rr *recordingRegistry) createdIDs() []string {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	ids := make([]string, 0, len(rr.created))
	for _, p := range rr.created {
		ids = append(ids, p.id)
	}
	return ids
}

// spec 构造一个 fake 类型的实例规格
func spec(id, name, token string) InstanceSpec {
	return InstanceSpec{ID: id, Type: "fake", Name: name, Config: map[string]string{"token": token}}
}

// noopHandler 不做任何处理的消息处理器
func noopHandler(context.Context, BotProvider, Message) {}

// providerByID 从全部运行中实例里按 ID 查找，未运行时返回 nil
func providerByID(m *Manager, id string) BotProvider {
	for _, p := range m.Providers() {
		if p.ID() == id {
			return p
		}
	}
	return nil
}

// waitStarted 等待实例的 Start 真正被调用，避免依赖 sleep
func waitStarted(t *testing.T, m *Manager, id string) *fakeProvider {
	t.Helper()

	provider := providerByID(m, id)
	require.NotNil(t, provider, "实例 %s 应处于运行中", id)
	p := provider.(*fakeProvider)
	select {
	case <-p.started:
	case <-time.After(2 * time.Second):
		t.Fatalf("实例 %s 的 Start 未被调用", id)
	}
	return p
}

// runningIDs 返回当前运行中实例的 ID 列表（已按 ID 排序）
func runningIDs(m *Manager) []string {
	statuses := m.Statuses()
	ids := make([]string, 0, len(statuses))
	for _, s := range statuses {
		ids = append(ids, s.ID)
	}
	return ids
}

// TestManager_StartAndStopAll 首批实例启动后可被统一停止
func TestManager_StartAndStopAll(t *testing.T) {
	rr := newRecordingRegistry()
	m := NewManager(rr.Registry, noopHandler)

	require.NoError(t, m.Start(context.Background(), []InstanceSpec{
		spec("a", "实例A", "t1"),
		spec("b", "实例B", "t2"),
	}))

	waitStarted(t, m, "a")
	waitStarted(t, m, "b")

	statuses := m.Statuses()
	require.Len(t, statuses, 2)
	assert.Equal(t, "a", statuses[0].ID, "状态列表应按 ID 排序")
	assert.Equal(t, "实例B", statuses[1].Name)

	m.StopAll()
	assert.Empty(t, m.Statuses(), "StopAll 后不应残留运行中实例")
}

// TestManager_Apply_Diff 差分热加载：新增/删除/未改动的实例不重建
func TestManager_Apply_Diff(t *testing.T) {
	rr := newRecordingRegistry()
	m := NewManager(rr.Registry, noopHandler)
	t.Cleanup(m.StopAll)

	require.NoError(t, m.Start(context.Background(), []InstanceSpec{
		spec("keep", "保留", "t1"),
		spec("drop", "待删", "t2"),
	}))
	keepBefore := waitStarted(t, m, "keep")
	waitStarted(t, m, "drop")

	require.NoError(t, m.Apply([]InstanceSpec{
		spec("keep", "保留", "t1"),
		spec("add", "新增", "t3"),
	}))

	assert.Equal(t, []string{"add", "keep"}, runningIDs(m), "drop 应被停止，add 应被启动")

	keepAfter := providerByID(m, "keep")
	require.NotNil(t, keepAfter)
	assert.Same(t, keepBefore, keepAfter, "配置未变的实例不应被重建")
	assert.Equal(t, []string{"keep", "drop", "add"}, rr.createdIDs(), "keep 只应被创建一次")
}

// TestManager_Apply_RebuildsOnConfigChange 凭据或显示名变化必须触发重建
func TestManager_Apply_RebuildsOnConfigChange(t *testing.T) {
	tests := []struct {
		name    string
		updated InstanceSpec
	}{
		{name: "token 变化", updated: spec("x", "实例X", "new-token")},
		{name: "显示名变化", updated: spec("x", "新名字", "t1")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := newRecordingRegistry()
			m := NewManager(rr.Registry, noopHandler)
			t.Cleanup(m.StopAll)

			require.NoError(t, m.Start(context.Background(), []InstanceSpec{spec("x", "实例X", "t1")}))
			before := waitStarted(t, m, "x")

			require.NoError(t, m.Apply([]InstanceSpec{tt.updated}))
			after := waitStarted(t, m, "x")

			assert.NotSame(t, before, after, "配置变化必须重建实例")
			assert.Equal(t, []string{"x", "x"}, rr.createdIDs())
		})
	}
}

// TestManager_Apply_PartialFailure 单实例失败不得拖垮其他实例
func TestManager_Apply_PartialFailure(t *testing.T) {
	rr := newRecordingRegistry("bad")
	m := NewManager(rr.Registry, noopHandler)
	t.Cleanup(m.StopAll)

	err := m.Start(context.Background(), []InstanceSpec{
		spec("good1", "正常1", "t1"),
		spec("bad", "坏的", "t2"),
		spec("good2", "正常2", "t3"),
	})

	require.Error(t, err, "存在失败实例时应返回聚合错误")
	assert.Contains(t, err.Error(), "1/3 个 Bot 实例启动失败")
	assert.Contains(t, err.Error(), "模拟实例 bad 认证失败")
	assert.Equal(t, []string{"good1", "good2"}, runningIDs(m), "其余实例必须照常运行")
}

// TestManager_HandlerReceivesSourceProvider 处理器必须拿到消息来源实例，避免多实例串台
func TestManager_HandlerReceivesSourceProvider(t *testing.T) {
	rr := newRecordingRegistry()

	type received struct {
		botID string
		msg   Message
	}
	got := make(chan received, 2)
	m := NewManager(rr.Registry, func(ctx context.Context, bot BotProvider, msg Message) {
		got <- received{botID: bot.ID(), msg: msg}
	})
	t.Cleanup(m.StopAll)

	require.NoError(t, m.Start(context.Background(), []InstanceSpec{
		spec("cs", "客服", "t1"),
		spec("alert", "告警", "t2"),
	}))

	alert := waitStarted(t, m, "alert")
	alert.handler(Message{BotID: "alert", BotType: "fake", UserID: "42", Content: "hi"})

	select {
	case r := <-got:
		assert.Equal(t, "alert", r.botID, "处理器必须收到消息来源实例")
		assert.Equal(t, "alert", r.msg.BotID)
	case <-time.After(2 * time.Second):
		t.Fatal("处理器未被调用")
	}
}

// TestNewManager_RejectsNilDeps 缺依赖属装配错误，必须立刻 panic
func TestNewManager_RejectsNilDeps(t *testing.T) {
	assert.Panics(t, func() { NewManager(nil, noopHandler) })
	assert.Panics(t, func() { NewManager(NewRegistry(), nil) })
}

// TestInstanceSpec_Fingerprint 指纹只随内容变化，与 map 遍历顺序无关
func TestInstanceSpec_Fingerprint(t *testing.T) {
	base := InstanceSpec{ID: "x", Type: "fake", Name: "n", Config: map[string]string{"token": "t", "channel": "c"}}
	reordered := InstanceSpec{ID: "x", Type: "fake", Name: "n", Config: map[string]string{"channel": "c", "token": "t"}}
	assert.Equal(t, base.fingerprint(), reordered.fingerprint(), "map 顺序不应影响指纹")

	changed := InstanceSpec{ID: "x", Type: "fake", Name: "n", Config: map[string]string{"token": "t2", "channel": "c"}}
	assert.NotEqual(t, base.fingerprint(), changed.fingerprint(), "凭据变化必须改变指纹")

	renamedID := InstanceSpec{ID: "y", Type: "fake", Name: "n", Config: base.Config}
	assert.Equal(t, base.fingerprint(), renamedID.fingerprint(), "ID 由 map 键承载，不参与指纹")
}
