package approval

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	approvalModel "micro-net-hub/internal/module/approval/model"
)

// EscapeText 转义用户可控文本, 避免破坏 Bot(HTML 模式)的消息排版.
//
// 先替换 & 再替换尖括号: strings.NewReplacer 单遍扫描, 不会对已替换结果二次转义.
func EscapeText(text string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(text)
}

// formatTime 审批相关文案的统一时间格式
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// metaGet 从审批单的认证上下文 JSON 中读取指定 key, 解析失败或缺失时返回空串
func metaGet(req *approvalModel.ApprovalRequest, key string) string {
	if req == nil || req.Meta == "" {
		return ""
	}
	meta := make(Meta)
	if err := json.Unmarshal([]byte(req.Meta), &meta); err != nil {
		return ""
	}
	return meta[key]
}

// remainingText 把剩余有效期描述为可读文案
func remainingText(expireAt, now time.Time) string {
	left := expireAt.Sub(now)
	if left <= 0 {
		return "已过期"
	}
	if left < time.Minute {
		return fmt.Sprintf("%d 秒", int(left.Seconds()))
	}
	return fmt.Sprintf("%d 分钟", int(left.Minutes()))
}

// noticeHTML 审批人收到的审批通知(HTML 格式).
func noticeHTML(req *approvalModel.ApprovalRequest, now time.Time, waitSeconds int) string {
	var sb strings.Builder
	sb.WriteString("🔐 <b>登录审批</b>\n")
	fmt.Fprintf(&sb, "申请人: %s", EscapeText(req.Username))
	if req.Nickname != "" {
		fmt.Fprintf(&sb, "(%s)", EscapeText(req.Nickname))
	}
	if sourceAddr := metaGet(req, MetaKeySourceAddr); sourceAddr != "" {
		fmt.Fprintf(&sb, "\n来源: %s", EscapeText(sourceAddr))
	}
	if sourceID := metaGet(req, MetaKeySourceID); sourceID != "" {
		fmt.Fprintf(&sb, " · 设备: %s", EscapeText(sourceID))
	}
	fmt.Fprintf(&sb, "\n时间: %s", formatTime(req.CreatedAt))
	fmt.Fprintf(&sb, "\n申请ID: <b>#%d</b> (有效期 %s)", req.ID, remainingText(req.ExpireAt, now))
	fmt.Fprintf(&sb, "\n\n用户正在等待连接(约 %d 秒); 若已超时, 批准后用户需重新连接方可生效.", waitSeconds)
	fmt.Fprintf(&sb, "\n\n/approve %d (通过)\n/reject %d [原因] (拒绝)", req.ID, req.ID)
	return sb.String()
}

// approverApprovedText 审批人操作通过后的回执
func approverApprovedText(req *approvalModel.ApprovalRequest, grantExpireAt time.Time) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "✅ 已通过 #%d (申请人 %s)", req.ID, EscapeText(req.Username))
	fmt.Fprintf(&sb, "\n放行有效期至 %s · 用户需重新登录", formatTime(grantExpireAt))
	return sb.String()
}

// approverRejectedText 审批人操作拒绝后的回执
func approverRejectedText(req *approvalModel.ApprovalRequest, cooldown time.Duration) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "❌ 已拒绝 #%d (申请人 %s)", req.ID, EscapeText(req.Username))
	if req.Reason != "" {
		fmt.Fprintf(&sb, "\n原因: %s", EscapeText(req.Reason))
	}
	fmt.Fprintf(&sb, "\n该用户 %d 秒内登录审批不会重复通知, 期间可重新申请但不会打扰审批人", int(cooldown.Seconds()))
	return sb.String()
}

// approverAlreadyDecidedText 申请单已被处理时的幂等回执
func approverAlreadyDecidedText(req *approvalModel.ApprovalRequest) string {
	decider := req.Decider
	if decider == "" {
		decider = "未知"
	}
	decidedAt := ""
	if req.DecidedAt != nil {
		decidedAt = " " + formatTime(*req.DecidedAt)
	}
	return fmt.Sprintf("ℹ️ #%d 已由 %s%s 处理: %s", req.ID, EscapeText(decider), decidedAt, req.StatusText())
}

// approverRequestGoneText 申请单已过期/已被清理时的提示
func approverRequestGoneText(id uint) string {
	return fmt.Sprintf("⚠️ 未找到待审批的申请 #%d (可能已过期或被清理)", id)
}

// applicantSubmittedText 申请人收到的"已提交审批"提示
func applicantSubmittedText(req *approvalModel.ApprovalRequest, now time.Time) string {
	return fmt.Sprintf("⏳ 已提交人工审批 #%d, 有效期 %s, 请稍候; 审批通过后请重新连接。",
		req.ID, remainingText(req.ExpireAt, now))
}

// applicantApprovedText 申请人收到的审批通过通知
func applicantApprovedText(req *approvalModel.ApprovalRequest, grantExpireAt time.Time) string {
	return fmt.Sprintf("✅ 你的登录申请 #%d 已通过, 请在 %s 前重新连接 (授权至 %s)。",
		req.ID, formatTime(grantExpireAt), formatTime(grantExpireAt))
}

// applicantRejectedText 申请人收到的审批拒绝通知
func applicantRejectedText(req *approvalModel.ApprovalRequest, cooldown time.Duration) string {
	text := fmt.Sprintf("❌ 你的登录申请 #%d 已被拒绝。", req.ID)
	if req.Reason != "" {
		text += fmt.Sprintf("\n原因: %s", EscapeText(req.Reason))
	}
	if cooldown > 0 {
		text += fmt.Sprintf("\n%d 秒内请勿重复申请。", int(cooldown.Seconds()))
	}
	return text
}

// pendingListText 待审批列表
func pendingListText(reqs []*approvalModel.ApprovalRequest, now time.Time) string {
	if len(reqs) == 0 {
		return "当前没有待审批申请"
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "待审批申请 %d 条:", len(reqs))
	for _, req := range reqs {
		name := req.Username
		if req.Nickname != "" {
			name = fmt.Sprintf("%s(%s)", req.Username, req.Nickname)
		}
		fmt.Fprintf(&sb, "\n#%d %s · %s · 剩余 %s",
			req.ID, EscapeText(name), EscapeText(metaGet(req, MetaKeySourceAddr)), remainingText(req.ExpireAt, now))
	}
	sb.WriteString("\n\n/approve <申请ID> (通过)\n/reject <申请ID> [原因] (拒绝)")
	return sb.String()
}

// helpText 审批机器人的帮助文案
func helpText() string {
	return strings.Join([]string{
		"登录审批机器人",
		"",
		"/pending (查看待审批申请)",
		"/approve <申请ID> [原因] (通过)",
		"/reject <申请ID> [原因] (拒绝)",
		"/grant <用户名> <分钟数> (应急放行)",
		"/id (查看自己的 chat id)",
		"/help (显示本帮助)",
	}, "\n")
}

// notAllowedText 非白名单会话的提示(附带自身标识, 便于自助配置)
func notAllowedText(chatID, userID string) string {
	return fmt.Sprintf("⛔ 你不在审批人白名单中, 无法执行审批操作。\n你的 chat id: %s\nuser id: %s\n请联系管理员将其加入 radius.approval.bot.approver-chat-ids。",
		chatID, userID)
}

// privateOnlyText 非私聊会话的提示
func privateOnlyText() string {
	return "⚠️ 审批操作仅支持私聊, 请在与机器人的私聊会话中执行。"
}

// rateLimitedText 触发频控的提示
func rateLimitedText(seconds int) string {
	return fmt.Sprintf("⚠️ 操作过于频繁, 请 %d 秒后再试。", seconds)
}

// unknownCommandText 无法识别的指令提示
func unknownCommandText() string {
	return "未识别的指令。\n\n" + helpText()
}
