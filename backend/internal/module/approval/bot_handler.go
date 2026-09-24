package approval

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"micro-net-hub/internal/bot"
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
)

// commandLimiter 审批指令频控器, 在包初始化时按默认值构造.
//
// 限值取自 radius.approval.bot.max-commands-per10s, 配置热更新后于下次调用生效
// (见 currentCommandLimiter)。用 atomic.Value 承载指针, 避免热更新时替换整个
// 限流器与 allow() 并发读取之间产生 data race.
var commandLimiter atomic.Value // 存储 *rateLimiter

func init() {
	commandLimiter.Store(newRateLimiter(5, 10*time.Second))
}

// currentCommandLimiter 按当前配置构造/复用频控器
func currentCommandLimiter() *rateLimiter {
	approval := configApproval()
	if approval == nil || approval.Bot.MaxCommandsPer10s <= 0 {
		return commandLimiter.Load().(*rateLimiter)
	}
	lim := commandLimiter.Load().(*rateLimiter)
	if lim.limit != approval.Bot.MaxCommandsPer10s {
		commandLimiter.Store(newRateLimiter(approval.Bot.MaxCommandsPer10s, 10*time.Second))
		return commandLimiter.Load().(*rateLimiter)
	}
	return lim
}

// BotHandler 处理来自 Bot 的审批指令, 适配 bot.MessageHandler 签名.
//
// 安全边界(按顺序):
//  1. 仅处理配置中参与审批的 Bot 实例;
//  2. 仅响应指令形态的消息, 忽略闲聊;
//  3. private-only 开启时仅接受私聊会话;
//  4. 仅接受审批人白名单(approver-chat-ids)中的会话;
//  5. 单会话指令频控.
func BotHandler(ctx context.Context, bot bot.BotProvider, msg bot.Message) {
	if !Enabled() {
		return
	}
	approval := configApproval()
	if approval == nil || !approval.Bot.Enable {
		return
	}
	if !instanceEnabled(approval, bot.ID()) {
		return
	}

	cmd, err := ParseCommand(msg.Content)
	if err != nil {
		// 指令格式错误: 回执用法提示(仅在稍后的白名单校验通过时才会真正发出)
		cmd = Command{Kind: CommandUnknown}
	}
	if cmd.Kind == CommandUnknown && err == nil {
		return
	}

	if PrivateOnly() && msg.ChatType != "" && msg.ChatType != "private" {
		_ = sendPlain(ctx, bot, msg.ChannelID, privateOnlyText()+"\n\n"+helpText())
		return
	}

	if allowed, left := currentCommandLimiter().allow(msg.ChannelID); !allowed {
		_ = sendPlain(ctx, bot, msg.ChannelID, rateLimitedText(left))
		return
	}

	if !IsApproverChat(msg.ChannelID) {
		global.Log.Warnf("拒绝非白名单会话的审批指令: instance=%s, chatID=%s, userID=%s, content=%q",
			bot.ID(), msg.ChannelID, msg.UserID, truncate(msg.Content, 100))
		_ = sendPlain(ctx, bot, msg.ChannelID, notAllowedText(msg.ChannelID, msg.UserID))
		return
	}

	if err != nil {
		_ = sendPlain(ctx, bot, msg.ChannelID, err.Error()+"\n\n"+helpText())
		return
	}

	decider := fmt.Sprintf("bot:%s:%s", bot.Type(), msg.ChannelID)
	channel := fmt.Sprintf("bot:%s", bot.Type())

	switch cmd.Kind {
	case CommandHelp:
		_ = sendPlain(ctx, bot, msg.ChannelID, helpText())
	case CommandID:
		_ = sendPlain(ctx, bot, msg.ChannelID,
			fmt.Sprintf("你的 chat id: %s\nuser id: %s", msg.ChannelID, msg.UserID))
	case CommandPending:
		handlePendingCommand(ctx, bot, msg.ChannelID, cmd)
	case CommandApprove:
		handleApproveCommand(ctx, bot, msg.ChannelID, cmd, decider, channel)
	case CommandReject:
		handleRejectCommand(ctx, bot, msg.ChannelID, cmd, decider, channel)
	case CommandGrant:
		handleGrantCommand(ctx, bot, msg.ChannelID, cmd, decider, channel)
	default:
		_ = sendPlain(ctx, bot, msg.ChannelID, unknownCommandText())
	}
}

// instanceEnabled 判断某个 Bbot 实例是否参与审批
func instanceEnabled(approval *config.RadiusApproval, instanceID string) bool {
	if len(approval.Bot.InstanceIDs) == 0 {
		// 未限定实例时, 任意已启用实例都可以承载审批
		return true
	}
	return containsFold(approval.Bot.InstanceIDs, instanceID)
}

// handlePendingCommand 处理 /pending
func handlePendingCommand(ctx context.Context, bot bot.BotProvider, chatID string, cmd Command) {
	reqs, err := ListPending(cmd.Limit)
	if err != nil {
		global.Log.Errorf("查询待审批申请失败: %v", err)
		_ = sendPlain(ctx, bot, chatID, "查询待审批申请失败, 请查看服务端日志")
		return
	}
	_ = sendPlain(ctx, bot, chatID, pendingListText(reqs, time.Now()))
}

// handleApproveCommand 处理 /approve
func handleApproveCommand(ctx context.Context, bot bot.BotProvider, chatID string, cmd Command, decider, channel string) {
	decision, err := Approve(cmd.RequestID, decider, channel, cmd.Reason)
	if err != nil {
		global.Log.Errorf("处理审批通过失败, requestID=%d: %v", cmd.RequestID, err)
		_ = sendPlain(ctx, bot, chatID, "处理审批失败, 请查看服务端日志")
		return
	}
	if decision == nil || decision.Request == nil {
		_ = sendPlain(ctx, bot, chatID, approverRequestGoneText(cmd.RequestID))
		return
	}
	if !decision.Applied {
		_ = sendPlain(ctx, bot, chatID, approverAlreadyDecidedText(decision.Request))
		return
	}

	grantExpireAt := time.Now().Add(approvalTimeouts().grantTTL)
	if decision.Grant != nil {
		grantExpireAt = decision.Grant.ExpireAt
	}

	_ = sendPlain(ctx, bot, chatID, approverApprovedText(decision.Request, grantExpireAt))
	NotifyApplicant(ctx, decision.Request.Username, applicantApprovedText(decision.Request, grantExpireAt))
}

// handleRejectCommand 处理 /reject
func handleRejectCommand(ctx context.Context, bot bot.BotProvider, chatID string, cmd Command, decider, channel string) {
	decision, err := Reject(cmd.RequestID, decider, channel, cmd.Reason)
	if err != nil {
		global.Log.Errorf("处理审批拒绝失败, requestID=%d: %v", cmd.RequestID, err)
		_ = sendPlain(ctx, bot, chatID, "处理审批失败, 请查看服务端日志")
		return
	}
	if decision == nil || decision.Request == nil {
		_ = sendPlain(ctx, bot, chatID, approverRequestGoneText(cmd.RequestID))
		return
	}
	if !decision.Applied {
		_ = sendPlain(ctx, bot, chatID, approverAlreadyDecidedText(decision.Request))
		return
	}

	cooldown := RejectCooldownLeft(decision.Request.Username)
	_ = sendPlain(ctx, bot, chatID, approverRejectedText(decision.Request, cooldown))
	NotifyApplicant(ctx, decision.Request.Username, applicantRejectedText(decision.Request, cooldown))
}

// handleGrantCommand 处理 /grant 应急放行
func handleGrantCommand(ctx context.Context, bot bot.BotProvider, chatID string, cmd Command, decider, channel string) {
	ttl := time.Duration(cmd.TTLMinutes) * time.Minute
	grant, err := GrantAccess(cmd.Username, decider, channel, ttl)
	if err != nil {
		global.Log.Errorf("应急放行失败, username=%s: %v", cmd.Username, err)
		_ = sendPlain(ctx, bot, chatID, "应急放行失败, 请查看服务端日志")
		return
	}

	_ = sendPlain(ctx, bot, chatID,
		fmt.Sprintf("✅ 已放行用户 %s, 有效期至 %s", grant.Username, formatTime(grant.ExpireAt)))
	NotifyApplicant(ctx, grant.Username,
		fmt.Sprintf("✅ 管理员已为你的登录应急放行, 请在 %s 前连接。", formatTime(grant.ExpireAt)))
}
