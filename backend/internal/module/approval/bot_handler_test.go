package approval

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"micro-net-hub/internal/bot"
	"micro-net-hub/internal/config"
	accountModel "micro-net-hub/internal/module/account/model"
	approvalModel "micro-net-hub/internal/module/approval/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeBotProvider 记录发送内容的假 Bot 实例, 实现 bot.BotProvider
type fakeBotProvider struct {
	id   string
	typ  string
	name string

	mu   sync.Mutex
	sent []fakeSent

	handler func(msg bot.Message)
}

// fakeSent 一次发送的记录
type fakeSent struct {
	chatID string
	text   string
	html   bool
}

func (f *fakeBotProvider) ID() string   { return f.id }
func (f *fakeBotProvider) Type() string { return f.typ }
func (f *fakeBotProvider) Name() string { return f.name }

func (f *fakeBotProvider) Start(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (f *fakeBotProvider) SendMessage(_ context.Context, toID string, text string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, fakeSent{chatID: toID, text: text})
	return nil
}

func (f *fakeBotProvider) SendHTMLMessage(ctx context.Context, toID string, html string) error {
	if err := f.SendMessage(ctx, toID, html); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent[len(f.sent)-1].html = true
	return nil
}

func (f *fakeBotProvider) SendTyping(context.Context, string) error { return nil }

func (f *fakeBotProvider) RegisterHandler(handler func(msg bot.Message)) { f.handler = handler }

// messages 返回已发送消息的快照
func (f *fakeBotProvider) messages() []fakeSent {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]fakeSent, len(f.sent))
	copy(out, f.sent)
	return out
}

// newFakeBotProvider 创建一个假的 telegram 实例
func newFakeBotProvider(id string) *fakeBotProvider {
	return &fakeBotProvider{id: id, typ: "telegram", name: "审批机器人"}
}

// approvalBotConfig 构造测试用的 Bot 审批配置
func approvalBotConfig(instanceID string, approverChatIDs []string) config.ApprovalBot {
	return config.ApprovalBot{
		Enable:            true,
		InstanceIDs:       []string{instanceID},
		ApproverChatIDs:   approverChatIDs,
		PrivateOnly:       true,
		MaxCommandsPer10s: 10,
	}
}

// setupHandlerEnv 准备带 Bot 审批配置的测试环境
func setupHandlerEnv(t *testing.T, approval *config.ApprovalConfig, providers ...*fakeBotProvider) {
	t.Helper()
	setupApprovalEnv(t, approval)

	all := make([]bot.BotProvider, len(providers))
	for i, p := range providers {
		all[i] = p
	}
	SetProviderLister(func() []bot.BotProvider {
		return all
	})
	t.Cleanup(func() { SetProviderLister(nil) })
}

// handlerApproval 审批开启 + 单实例 + 指定白名单的默认配置
func handlerApproval(instanceID string, approverChatIDs []string) *config.ApprovalConfig {
	return &config.ApprovalConfig{
		Enable:            true,
		WaitSeconds:       1,
		PendingTTLSeconds: 120,
		GrantTTLMinutes:   30,
		Bot:               approvalBotConfig(instanceID, approverChatIDs),
	}
}

// TestBotHandler_IgnoresChitChat 非指令消息不产生任何响应
func TestBotHandler_IgnoresChitChat(t *testing.T) {
	b := newFakeBotProvider("bot1")
	setupHandlerEnv(t, handlerApproval("bot1", []string{"100"}), b)

	BotHandler(context.Background(), b, bot.Message{
		BotID: "bot1", BotType: "telegram", UserID: "9", ChannelID: "100",
		ChatType: "private", Content: "早上好呀",
	})
	assert.Empty(t, b.messages(), "闲聊不应得到回复")
}

// TestBotHandler_RejectsGroupChat private-only 生效时群聊指令被拒绝
func TestBotHandler_RejectsGroupChat(t *testing.T) {
	b := newFakeBotProvider("bot1")
	setupHandlerEnv(t, handlerApproval("bot1", []string{"100"}), b)

	BotHandler(context.Background(), b, bot.Message{
		BotID: "bot1", BotType: "telegram", UserID: "9", ChannelID: "100",
		ChatType: "group", Content: "/pending",
	})

	msgs := b.messages()
	require.Len(t, msgs, 1)
	assert.Contains(t, msgs[0].text, "仅支持私聊")
	assert.Equal(t, "100", msgs[0].chatID)
}

// TestBotHandler_RejectsNonApprover 非白名单会话无法执行审批并得到自助提示
func TestBotHandler_RejectsNonApprover(t *testing.T) {
	b := newFakeBotProvider("bot1")
	setupHandlerEnv(t, handlerApproval("bot1", []string{"100"}), b)

	req, _, err := CreateOrReuseRequest(&accountModel.User{Username: "alice"}, Meta{})
	require.NoError(t, err)

	contents := []string{"/pending", fmt.Sprintf("/approve %d 同意", req.ID), "/grant alice 10"}
	for _, content := range contents {
		BotHandler(context.Background(), b, bot.Message{
			BotID: "bot1", BotType: "telegram", UserID: "42", ChannelID: "999",
			ChatType: "private", Content: content,
		})
	}

	sent := b.messages()
	require.Len(t, sent, len(contents))
	for _, msg := range sent {
		assert.Contains(t, msg.text, "不在审批人白名单")
		assert.Contains(t, msg.text, "999", "提示应带出会话 ID 便于自助配置")
	}

	// 申请单必须仍为待审批状态
	pending, err := approvalModel.FindApprovalRequestByID(req.ID)
	require.NoError(t, err)
	require.NotNil(t, pending)
	assert.Equal(t, approvalModel.RequestStatusPending, pending.Status)
}

// TestBotHandler_SkipsOtherInstances 未参与审批的 Bot 实例消息被忽略
func TestBotHandler_SkipsOtherInstances(t *testing.T) {
	b := newFakeBotProvider("bot1")
	other := newFakeBotProvider("bot2")
	setupHandlerEnv(t, handlerApproval("bot1", []string{"100"}), b, other)

	BotHandler(context.Background(), other, bot.Message{
		BotID: "bot2", BotType: "telegram", UserID: "42", ChannelID: "100",
		ChatType: "private", Content: "/pending",
	})
	assert.Empty(t, other.messages())
}

// TestBotHandler_DisabledWithoutConfig 审批未开启或 Bot 审批未启用时完全静默
func TestBotHandler_DisabledWithoutConfig(t *testing.T) {
	b := newFakeBotProvider("bot1")

	disabled := handlerApproval("bot1", []string{"100"})
	disabled.Enable = false
	setupHandlerEnv(t, disabled, b)
	BotHandler(context.Background(), b, bot.Message{
		BotID: "bot1", ChannelID: "100", ChatType: "private", Content: "/pending",
	})
	assert.Empty(t, b.messages())

	botDisabled := handlerApproval("bot1", []string{"100"})
	botDisabled.Bot.Enable = false
	setupHandlerEnv(t, botDisabled, b)
	BotHandler(context.Background(), b, bot.Message{
		BotID: "bot1", ChannelID: "100", ChatType: "private", Content: "/pending",
	})
	assert.Empty(t, b.messages())
}

// TestBotHandler_CommandID 白名单用户可用 /id 自助查询会话标识
func TestBotHandler_CommandID(t *testing.T) {
	b := newFakeBotProvider("bot1")
	setupHandlerEnv(t, handlerApproval("bot1", []string{"100"}), b)

	BotHandler(context.Background(), b, bot.Message{
		BotID: "bot1", UserID: "42", ChannelID: "100", ChatType: "private", Content: "/id",
	})

	msgs := b.messages()
	require.Len(t, msgs, 1)
	assert.Contains(t, msgs[0].text, "100")
	assert.Contains(t, msgs[0].text, "42")
}

// TestBotHandler_CommandHelpAndBadArgs 帮助与非法参数的用法提示
func TestBotHandler_CommandHelpAndBadArgs(t *testing.T) {
	b := newFakeBotProvider("bot1")
	setupHandlerEnv(t, handlerApproval("bot1", []string{"100"}), b)

	ctx := context.Background()
	contents := []string{"/help", "/approve", "/approve abc"}
	for _, content := range contents {
		BotHandler(ctx, b, bot.Message{
			BotID: "bot1", UserID: "42", ChannelID: "100", ChatType: "private", Content: content,
		})
	}

	msgs := b.messages()
	require.Len(t, msgs, len(contents))
	assert.Contains(t, msgs[0].text, "/grant <用户名> <分钟数>")
	assert.Contains(t, msgs[1].text, "用法: /approve <申请ID>")
	assert.Contains(t, msgs[2].text, "申请ID 非法")
	assert.Contains(t, msgs[2].text, "/pending", "非法参数应附带帮助文案")
}

// TestBotHandler_ApproveFlow 白名单会话 /approve 后签发凭证并回执
func TestBotHandler_ApproveFlow(t *testing.T) {
	b := newFakeBotProvider("bot1")
	setupHandlerEnv(t, handlerApproval("bot1", []string{"100"}), b)

	req, _, err := CreateOrReuseRequest(&accountModel.User{Username: "alice"}, Meta{})
	require.NoError(t, err)

	ctx := context.Background()
	BotHandler(ctx, b, bot.Message{
		BotID: "bot1", UserID: "42", ChannelID: "100", ChatType: "private",
		Content: fmt.Sprintf("/approve %d 紧急上线", req.ID),
	})

	msgs := b.messages()
	require.Len(t, msgs, 1)
	assert.Contains(t, msgs[0].text, fmt.Sprintf("已通过 #%d", req.ID))

	grant, err := FindActiveGrant("alice")
	require.NoError(t, err)
	require.NotNil(t, grant)

	decided, err := approvalModel.FindApprovalRequestByID(req.ID)
	require.NoError(t, err)
	require.NotNil(t, decided)
	assert.Equal(t, approvalModel.RequestStatusApproved, decided.Status)
	assert.Equal(t, "bot:telegram:100", decided.Decider)
	assert.Equal(t, "bot:telegram", decided.DeciderChannel)
	assert.Equal(t, "紧急上线", decided.Reason)

	// 重复审批为幂等, 回执提示已由首位审批人处理
	BotHandler(ctx, b, bot.Message{
		BotID: "bot1", UserID: "42", ChannelID: "100", ChatType: "private",
		Content: fmt.Sprintf("/approve %d", req.ID),
	})

	msgs = b.messages()
	require.Len(t, msgs, 2)
	assert.Contains(t, msgs[1].text, "已由 bot:telegram:100")
}

// TestBotHandler_RejectFlow 拒绝后进入冷却并回执
func TestBotHandler_RejectFlow(t *testing.T) {
	b := newFakeBotProvider("bot1")
	setupHandlerEnv(t, handlerApproval("bot1", []string{"100"}), b)

	req, _, err := CreateOrReuseRequest(&accountModel.User{Username: "alice"}, Meta{})
	require.NoError(t, err)

	BotHandler(context.Background(), b, bot.Message{
		BotID: "bot1", UserID: "42", ChannelID: "100", ChatType: "private",
		Content: fmt.Sprintf("/reject %d 非工作时间", req.ID),
	})

	msgs := b.messages()
	require.Len(t, msgs, 1)
	assert.Contains(t, msgs[0].text, fmt.Sprintf("已拒绝 #%d", req.ID))
	assert.Contains(t, msgs[0].text, "非工作时间")
	assert.Positive(t, RejectCooldownLeft("alice").Seconds())

	// 拒绝后不签发凭证
	grant, err := FindActiveGrant("alice")
	require.NoError(t, err)
	assert.Nil(t, grant)
}

// TestBotHandler_GrantFlow 应急放行指令签发凭证
func TestBotHandler_GrantFlow(t *testing.T) {
	b := newFakeBotProvider("bot1")
	setupHandlerEnv(t, handlerApproval("bot1", []string{"100"}), b)

	BotHandler(context.Background(), b, bot.Message{
		BotID: "bot1", UserID: "42", ChannelID: "100", ChatType: "private",
		Content: "/grant alice 15",
	})

	msgs := b.messages()
	require.Len(t, msgs, 1)
	assert.Contains(t, msgs[0].text, "已放行用户 alice")

	grant, err := FindActiveGrant("alice")
	require.NoError(t, err)
	require.NotNil(t, grant)
	assert.WithinDuration(t, time.Now().Add(15*time.Minute), grant.ExpireAt, time.Minute)
}

// TestBotHandler_PendingList /pending 返回待审批清单
func TestBotHandler_PendingList(t *testing.T) {
	b := newFakeBotProvider("bot1")
	setupHandlerEnv(t, handlerApproval("bot1", []string{"100"}), b)

	ctx := context.Background()
	BotHandler(ctx, b, bot.Message{
		BotID: "bot1", ChannelID: "100", ChatType: "private", Content: "/pending",
	})

	req, _, err := CreateOrReuseRequest(&accountModel.User{Username: "alice"}, Meta{})
	require.NoError(t, err)
	BotHandler(ctx, b, bot.Message{
		BotID: "bot1", ChannelID: "100", ChatType: "private", Content: "/pending 5",
	})

	msgs := b.messages()
	require.Len(t, msgs, 2)
	assert.Equal(t, "当前没有待审批申请", msgs[0].text)
	assert.Contains(t, msgs[1].text, fmt.Sprintf("#%d alice", req.ID))
}

// TestBotHandler_RateLimited 超出单会话指令频控后被限流
func TestBotHandler_RateLimited(t *testing.T) {
	approval := handlerApproval("bot1", []string{"100"})
	approval.Bot.MaxCommandsPer10s = 2

	b := newFakeBotProvider("bot1")
	setupHandlerEnv(t, approval, b)

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		BotHandler(ctx, b, bot.Message{
			BotID: "bot1", ChannelID: "100", ChatType: "private", Content: "/pending",
		})
	}

	msgs := b.messages()
	require.Len(t, msgs, 3)
	assert.Contains(t, msgs[2].text, "操作过于频繁", "第 3 条指令应被限流")
}

// TestNotifyApprovers 审批通知以 HTML 私聊方式发给全部白名单会话, 并回执申请人
func TestNotifyApprovers(t *testing.T) {
	approval := handlerApproval("bot1", []string{"100", "200"})
	approval.Bot.ApplicantNotify = true
	approval.Bot.ApplicantMap = map[string]string{"alice": "300"}

	bot := newFakeBotProvider("bot1")
	setupHandlerEnv(t, approval, bot)

	req, _, err := CreateOrReuseRequest(&accountModel.User{Username: "alice", Nickname: "艾丽斯"},
		Meta{MetaKeySourceAddr: "10.0.0.1", MetaKeySourceID: "ocserv-1"})
	require.NoError(t, err)

	NotifyApprovers(context.Background(), req)

	sent := map[string]fakeSent{}
	for _, msg := range bot.messages() {
		sent[msg.chatID] = msg
	}

	require.Contains(t, sent, "100")
	require.Contains(t, sent, "200")
	assert.True(t, sent["100"].html, "审批通知应优先使用 HTML 格式")
	assert.Contains(t, sent["100"].text, fmt.Sprintf("#%d", req.ID))
	assert.Contains(t, sent["100"].text, "alice")
	assert.Contains(t, sent["100"].text, "ocserv-1")

	require.Contains(t, sent, "300", "申请人应收到已提交回执")
	assert.Contains(t, sent["300"].text, "已提交人工审批")

	// 通知成功后更新通知时间, 避免重复打扰审批人
	latest, err := approvalModel.FindApprovalRequestByID(req.ID)
	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.NotNil(t, latest.LastNoticeAt)
}

// TestNotifyApplicant_Skipped 关闭回执或未配置会话映射时不发送
func TestNotifyApplicant_Skipped(t *testing.T) {
	approval := handlerApproval("bot1", []string{"100"})
	approval.Bot.ApplicantMap = map[string]string{"alice": "300"}

	bot := newFakeBotProvider("bot1")
	setupHandlerEnv(t, approval, bot)

	NotifyApplicant(context.Background(), "alice", "结果通知")
	assert.Empty(t, bot.messages(), "未开启 applicant-notify 时不应发送")

	approval.Bot.ApplicantNotify = true
	setupHandlerEnv(t, approval, bot)
	NotifyApplicant(context.Background(), "bob", "结果通知")
	assert.Empty(t, bot.messages(), "未配置会话映射的用户应静默跳过")
}
