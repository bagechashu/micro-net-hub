package approval

import (
	"context"
	"strings"
	"time"

	"micro-net-hub/internal/bot"
	"micro-net-hub/internal/global"
	approvalModel "micro-net-hub/internal/module/approval/model"
)

// ProviderLister 列出全部运行中的实例, 由装配层注入.
//
// 用函数注入而非让本包直接依赖全局 Bot 管理器, 使审批模块只依赖 bot.BotProvider 接口,
// 便于测试时替换为假实现. 按 ID 挑选实例由 approvalProviders 在内存中查表完成,
// 无需再单独注入按 ID 解析的 resolver.
type ProviderLister func() []bot.BotProvider

// providerLister 审批通知使用的 Bot 实例列表器
var providerLister ProviderLister

// SetProviderLister 注入 Bot 实例列表器(进程启动时调用一次)
func SetProviderLister(lister ProviderLister) {
	providerLister = lister
}

// approvalProviders 返回参与审批且当前可用的 Bot 实例.
//
// 先列出全部运行中实例, 再按 instance-ids 过滤; instance-ids 为空表示不限定实例,
// 使用全部运行中的实例(与 instanceEnabled 语义一致).
func approvalProviders() []bot.BotProvider {
	approval := configApproval()
	if approval == nil || providerLister == nil {
		return nil
	}

	all := providerLister()
	if len(approval.Bot.InstanceIDs) == 0 {
		// 未限定实例时广播到全部运行中实例
		return all
	}

	byID := make(map[string]bot.BotProvider, len(all))
	for _, p := range all {
		byID[p.ID()] = p
	}

	providers := make([]bot.BotProvider, 0, len(approval.Bot.InstanceIDs))
	for _, id := range approval.Bot.InstanceIDs {
		id = strings.TrimSpace(id)
		provider, ok := byID[id]
		if !ok {
			global.Log.Warnf("审批配置引用的 Bot 实例 [%s] 未在运行, 已跳过", id)
			continue
		}
		providers = append(providers, provider)
	}
	return providers
}

// ApproverChatIDs 返回审批人会话白名单(去除空白项)
func ApproverChatIDs() []string {
	approval := configApproval()
	if approval == nil {
		return nil
	}

	out := make([]string, 0, len(approval.Bot.ApproverChatIDs))
	for _, id := range approval.Bot.ApproverChatIDs {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// IsApproverChat 判断某个会话是否在审批人白名单内
func IsApproverChat(chatID string) bool {
	return containsFold(ApproverChatIDs(), chatID)
}

// PrivateOnly 判断是否仅允许私聊会话审批(默认 true)
func PrivateOnly() bool {
	approval := configApproval()
	if approval == nil {
		return true
	}
	return approval.Bot.PrivateOnly
}

// NotifyApprovers 向全部审批人发送审批通知, 并回执申请人"已提交"
func NotifyApprovers(ctx context.Context, req *approvalModel.ApprovalRequest) {
	if req == nil {
		return
	}

	approval := configApproval()
	if approval == nil || !approval.Bot.Enable {
		global.Log.Warnf("审批申请 #%d 已创建, 但 Bot 审批未启用, 需管理员通过其它方式处理", req.ID)
		return
	}

	chatIDs := ApproverChatIDs()
	if len(chatIDs) == 0 {
		global.Log.Warnf("审批申请 #%d 无法通知审批人: radius.approval.bot.approver-chat-ids 未配置", req.ID)
		return
	}

	providers := approvalProviders()
	if len(providers) == 0 {
		global.Log.Warnf("审批申请 #%d 无法通知审批人: 没有正在运行的审批 Bot 实例 (instance-ids=%v)",
			req.ID, approval.Bot.InstanceIDs)
		return
	}

	now := time.Now()
	text := noticeHTML(req, now, int(approvalTimeouts().wait.Seconds()))

	sent := 0
	for _, provider := range providers {
		for _, chatID := range chatIDs {
			if err := sendHTML(ctx, provider, chatID, text); err != nil {
				global.Log.Errorf("发送审批通知失败, instance=%s, chatID=%s: %v", provider.ID(), chatID, err)
				continue
			}
			sent++
		}
	}

	if sent > 0 {
		MarkNoticed(req, now)
		global.Log.Infof("审批申请 #%d 已通知 %d 个审批会话", req.ID, sent)
	}

	NotifyApplicant(ctx, req.Username, applicantSubmittedText(req, now))
}

// NotifyApplicant 向申请人回执审批进度或结果.
//
// 申请人会话通过 applicant-map 配置或申请单上的会话映射获取; 未配置时静默跳过
// (例如申请人从未与机器人建立会话, 此时无法主动私聊).
func NotifyApplicant(ctx context.Context, username, text string) {
	if text == "" {
		return
	}

	approval := configApproval()
	if approval == nil || !approval.Bot.Enable || !approval.Bot.ApplicantNotify {
		return
	}

	chatID := applicantChatID(username)
	if chatID == "" {
		global.Log.Debugf("未配置用户 %s 的 Bot 会话, 跳过申请人回执", username)
		return
	}

	for _, provider := range approvalProviders() {
		if err := sendPlain(ctx, provider, chatID, text); err != nil {
			global.Log.Errorf("发送申请人回执失败, instance=%s, chatID=%s: %v", provider.ID(), chatID, err)
		}
	}
}

// applicantChatID 查询申请人的 Bot 会话 ID
func applicantChatID(username string) string {
	approval := configApproval()
	if approval == nil || len(approval.Bot.ApplicantMap) == 0 {
		return ""
	}
	for name, chatID := range approval.Bot.ApplicantMap {
		if strings.EqualFold(strings.TrimSpace(name), username) {
			return strings.TrimSpace(chatID)
		}
	}
	return ""
}

// sendHTML 发送 HTML 消息, 失败时降级为纯文本, 避免因排版实体不受支持而丢通知
func sendHTML(ctx context.Context, provider bot.BotProvider, chatID, html string) error {
	err := provider.SendHTMLMessage(ctx, chatID, html)
	if err == nil {
		return nil
	}
	if plainErr := provider.SendMessage(ctx, chatID, stripHTML(html)); plainErr == nil {
		global.Log.Warnf("Bot 实例 [%s] 发送 HTML 消息失败, 已降级为纯文本: %v", provider.ID(), err)
		return nil
	}
	return err
}

// sendPlain 发送纯文本消息
func sendPlain(ctx context.Context, provider bot.BotProvider, chatID, text string) error {
	return provider.SendMessage(ctx, chatID, text)
}

// htmlStripper 纯文本降级时使用的标签替换器
var htmlStripper = strings.NewReplacer(
	"<b>", "", "</b>", "",
	"<i>", "", "</i>", "",
	"<u>", "", "</u>", "",
	"<s>", "", "</s>", "",
	"<code>", "", "</code>", "",
	"<pre>", "", "</pre>", "",
	"&lt;", "<", "&gt;", ">", "&amp;", "&",
)

// stripHTML 去掉消息中的 HTML 标签与实体(仅用于降级发送)
func stripHTML(text string) string {
	return htmlStripper.Replace(text)
}
