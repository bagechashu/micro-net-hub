package bot

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeProvider 一个不依赖外部网络的 BotProvider 实现
type fakeProvider struct {
	id      string
	botType string
	name    string
	handler func(Message)
	started chan struct{}
	sent    []string
}

func (f *fakeProvider) ID() string   { return f.id }
func (f *fakeProvider) Type() string { return f.botType }
func (f *fakeProvider) Name() string { return f.name }

func (f *fakeProvider) Start(ctx context.Context) error {
	close(f.started)
	<-ctx.Done()
	return ctx.Err()
}

func (f *fakeProvider) SendMessage(ctx context.Context, toID string, text string) error {
	f.sent = append(f.sent, text)
	return nil
}

func (f *fakeProvider) SendHTMLMessage(ctx context.Context, toID string, html string) error {
	return f.SendMessage(ctx, toID, html)
}

func (f *fakeProvider) SendTyping(ctx context.Context, toID string) error { return nil }

func (f *fakeProvider) RegisterHandler(handler func(Message)) { f.handler = handler }

// fakeSpec 测试用类型规格：一个必填且敏感的 token，一个可选的非敏感项
var fakeSpec = TypeSpec{
	Type:          "fake",
	Label:         "Fake",
	RequiredKeys:  []string{"token"},
	OptionalKeys:  []string{"channel"},
	SensitiveKeys: []string{"token"},
}

// newFakeRegistry 构造只注册了 fake 类型的注册表
func newFakeRegistry() *Registry {
	r := NewRegistry()
	r.Register(fakeSpec, func(id, name string, config map[string]string) (BotProvider, error) {
		return &fakeProvider{
			id:      id,
			botType: "fake",
			name:    name,
			started: make(chan struct{}),
		}, nil
	})
	return r
}

// TestRegistry_Create 覆盖已注册/未注册/缺必填项三类创建路径
func TestRegistry_Create(t *testing.T) {
	tests := []struct {
		name    string
		botType string
		config  map[string]string
		wantErr string
	}{
		{name: "已注册类型且配置完整", botType: "fake", config: map[string]string{"token": "t"}},
		{name: "未注册类型直接报错", botType: "feishu", config: map[string]string{"token": "t"},
			wantErr: `未注册的 Bot 类型 "feishu"`},
		{name: "缺必填配置项", botType: "fake", config: map[string]string{},
			wantErr: "缺少必填配置项: [token]"},
		{name: "必填项为空字符串等同缺失", botType: "fake", config: map[string]string{"token": ""},
			wantErr: "缺少必填配置项: [token]"},
	}

	registry := newFakeRegistry()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := registry.Create(tt.botType, "inst1", "实例1", tt.config)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, provider)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "inst1", provider.ID())
			assert.Equal(t, "fake", provider.Type())
			assert.Equal(t, "实例1", provider.Name())
		})
	}
}

// TestRegistry_ValidateConfig 写配置前的类型专属校验
func TestRegistry_ValidateConfig(t *testing.T) {
	registry := newFakeRegistry()

	require.NoError(t, registry.ValidateConfig("fake", map[string]string{"token": "t"}))
	assert.ErrorContains(t, registry.ValidateConfig("fake", nil), "缺少必填配置项")
	assert.ErrorContains(t, registry.ValidateConfig("wecom", map[string]string{}), "未注册的 Bot 类型")
}

// TestRegistry_SpecsAndTypes 类型元数据供前端渲染表单，必须稳定排序
func TestRegistry_SpecsAndTypes(t *testing.T) {
	r := NewRegistry()
	r.Register(TypeSpec{Type: "zeta"}, func(string, string, map[string]string) (BotProvider, error) {
		return nil, fmt.Errorf("unused")
	})
	r.Register(fakeSpec, func(string, string, map[string]string) (BotProvider, error) {
		return nil, fmt.Errorf("unused")
	})

	assert.Equal(t, []string{"fake", "zeta"}, r.SupportedTypes(), "类型列表应按字典序")

	specs := r.Specs()
	require.Len(t, specs, 2)
	assert.Equal(t, "fake", specs[0].Type)
	assert.Equal(t, []string{"token"}, specs[0].SensitiveKeys)

	spec, ok := r.Spec("fake")
	require.True(t, ok)
	assert.True(t, spec.IsSensitive("token"))
	assert.False(t, spec.IsSensitive("channel"))

	_, ok = r.Spec("missing")
	assert.False(t, ok, "未注册类型不应返回规格")
}

// TestDefaultRegistry_HasTelegram 内置注册表必须自带 telegram，且 token 被标为敏感
func TestDefaultRegistry_HasTelegram(t *testing.T) {
	spec, ok := DefaultRegistry().Spec("telegram")
	require.True(t, ok, "telegram 必须是内置类型")
	assert.Equal(t, []string{"token"}, spec.RequiredKeys)
	assert.True(t, spec.IsSensitive("token"), "Bot Token 必须被判定为敏感项")
}

// TestRegistry_RegisterRejectsInvalid 非法注册属编程错误，必须 panic 而非静默忽略
func TestRegistry_RegisterRejectsInvalid(t *testing.T) {
	r := NewRegistry()

	assert.Panics(t, func() {
		r.Register(TypeSpec{Type: ""}, func(string, string, map[string]string) (BotProvider, error) {
			return nil, nil
		})
	}, "空类型标识必须 panic")

	assert.Panics(t, func() {
		r.Register(TypeSpec{Type: "x"}, nil)
	}, "nil 工厂函数必须 panic")
}
