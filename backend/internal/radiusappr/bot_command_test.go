package radiusappr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseCommand 覆盖各审批指令的解析与非法参数
func TestParseCommand(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    Command
		wantErr bool
	}{
		{name: "非指令文本被忽略", text: "你好, 在吗", want: Command{}},
		{name: "空文本被忽略", text: "   ", want: Command{}},
		{name: "通过指令带原因", text: "/approve 12 紧急上线",
			want: Command{Kind: CommandApprove, RequestID: 12, Reason: "紧急上线"}},
		{name: "兼容 # 前缀的申请ID", text: "/approve #12",
			want: Command{Kind: CommandApprove, RequestID: 12}},
		{name: "群聊中的 @botname 后缀", text: "/reject@vpn_approval_bot 7 非工作时间",
			want: Command{Kind: CommandReject, RequestID: 7, Reason: "非工作时间"}},
		{name: "拒绝指令缺少申请ID", text: "/reject", wantErr: true},
		{name: "申请ID 非数字", text: "/approve abc", wantErr: true},
		{name: "申请ID 为 0", text: "/approve 0", wantErr: true},
		{name: "待审批列表默认条数", text: "/pending",
			want: Command{Kind: CommandPending, Limit: defaultPendingLimit}},
		{name: "待审批列表指定条数", text: "/pending 3",
			want: Command{Kind: CommandPending, Limit: 3}},
		{name: "待审批列表非法条数回落默认值", text: "/pending x",
			want: Command{Kind: CommandPending, Limit: defaultPendingLimit}},
		{name: "帮助", text: "/help", want: Command{Kind: CommandHelp}},
		{name: "start 等同帮助", text: "/start", want: Command{Kind: CommandHelp}},
		{name: "查看会话标识", text: "/id", want: Command{Kind: CommandID}},
		{name: "应急放行带分钟数", text: "/grant alice 120",
			want: Command{Kind: CommandGrant, Username: "alice", TTLMinutes: 120}},
		{name: "应急放行缺少用户名", text: "/grant", wantErr: true},
		{name: "未知指令被忽略", text: "/unknown", want: Command{}},
		{name: "大小写不敏感", text: "/APPROVE 5",
			want: Command{Kind: CommandApprove, RequestID: 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseCommand(tt.text)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, cmd)
		})
	}
}

// TestParseCommand_TruncatesLongInput 超长原因与用户名被截断, 保护数据库字段长度
func TestParseCommand_TruncatesLongInput(t *testing.T) {
	longReason := ""
	for i := 0; i < 300; i++ {
		longReason += "原"
	}

	cmd, err := ParseCommand("/reject 1 " + longReason)
	require.NoError(t, err)
	assert.Len(t, []rune(cmd.Reason), 200)
}

// TestRateLimiter 固定窗口频控: 达上限后拒绝, 窗口过期后恢复
func TestRateLimiter(t *testing.T) {
	limiter := newRateLimiter(2, 50*time.Millisecond)

	allowed, _ := limiter.allow("chat-1")
	assert.True(t, allowed)
	allowed, _ = limiter.allow("chat-1")
	assert.True(t, allowed)

	allowed, retryAfter := limiter.allow("chat-1")
	assert.False(t, allowed)
	assert.Positive(t, retryAfter)

	// 其它会话的额度互相独立
	allowed, _ = limiter.allow("chat-2")
	assert.True(t, allowed)

	time.Sleep(60 * time.Millisecond)
	allowed, _ = limiter.allow("chat-1")
	assert.True(t, allowed, "窗口过期后应重新放行")
}

// TestRateLimiter_Defaults 非法参数回落默认值
func TestRateLimiter_Defaults(t *testing.T) {
	limiter := newRateLimiter(0, 0)
	assert.Equal(t, 5, limiter.limit)
	assert.Equal(t, 10*time.Second, limiter.window)
}
