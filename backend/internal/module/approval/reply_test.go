package approval

import (
	"testing"
	"time"

	approvalModel "micro-net-hub/internal/module/approval/model"

	"github.com/stretchr/testify/assert"
)

// TestEscapeText 用户可控内容必须转义, 避免破坏 HTML 排版
func TestEscapeText(t *testing.T) {
	assert.Equal(t, "&lt;b&gt;x&lt;/b&gt;", EscapeText("<b>x</b>"))
	assert.Equal(t, "a &amp; b", EscapeText("a & b"))
	assert.Equal(t, "&amp;lt;", EscapeText("&lt;"), "已转义内容不应被还原成可执行标签")
}

// TestNoticeHTML_EscapesUserInput 审批通知中的用户输入必须被转义
func TestNoticeHTML_EscapesUserInput(t *testing.T) {
	now := time.Now()
	req := &approvalModel.ApprovalRequest{
		Username: "alice<b>",
		Nickname: "<script>alert(1)</script>",
		Meta:     `{"sourceAddr":"10.0.0.1"}`,
		ExpireAt: now.Add(2 * time.Minute),
	}
	req.ID = 42

	text := noticeHTML(req, now, 6)
	assert.Contains(t, text, "alice&lt;b&gt;")
	assert.Contains(t, text, "&lt;script&gt;")
	assert.Contains(t, text, "#42")
	assert.Contains(t, text, "/approve 42")
	assert.Contains(t, text, "/reject 42")
	assert.NotContains(t, text, "<script>")
}

// TestStripHTML 降级发送时标签与实体被还原
func TestStripHTML(t *testing.T) {
	assert.Equal(t, "加粗 文本", stripHTML("<b>加粗</b> <i>文本</i>"))
	assert.Equal(t, "a < b & c", stripHTML("a &lt; b &amp; c"))
}

// TestPendingListText 空列表与多条目列表
func TestPendingListText(t *testing.T) {
	now := time.Now()
	assert.Equal(t, "当前没有待审批申请", pendingListText(nil, now))

	reqs := []*approvalModel.ApprovalRequest{
		{Username: "alice", Nickname: "艾丽斯", Meta: `{"sourceAddr":"10.0.0.1"}`, ExpireAt: now.Add(time.Minute)},
		{Username: "bob", Meta: `{"sourceAddr":"10.0.0.2"}`, ExpireAt: now.Add(-time.Minute)},
	}
	reqs[0].ID = 1
	reqs[1].ID = 2

	text := pendingListText(reqs, now)
	assert.Contains(t, text, "#1 alice(艾丽斯)")
	assert.Contains(t, text, "剩余 1 分钟")
	assert.Contains(t, text, "剩余 已过期")
	assert.Contains(t, text, "/approve")
}

// TestRemainingText 剩余有效期文案
func TestRemainingText(t *testing.T) {
	now := time.Now()
	assert.Equal(t, "已过期", remainingText(now.Add(-time.Second), now))
	assert.Equal(t, "30 秒", remainingText(now.Add(30*time.Second), now))
	assert.Equal(t, "5 分钟", remainingText(now.Add(5*time.Minute), now))
}

// TestHelpAndNotAllowedText 帮助与白名单提示需附带自助排障信息
func TestHelpAndNotAllowedText(t *testing.T) {
	help := helpText()
	assert.Contains(t, help, "/approve <申请ID>")
	assert.Contains(t, help, "/id")

	text := notAllowedText("12345", "6789")
	assert.Contains(t, text, "12345")
	assert.Contains(t, text, "6789")
	assert.Contains(t, text, "approver-chat-ids")
}

// TestApproverReplyTexts 审批人回执文案
func TestApproverReplyTexts(t *testing.T) {
	now := time.Now()
	decided := now
	req := &approvalModel.ApprovalRequest{
		Username:  "alice",
		Reason:    "非工作时间",
		DecidedAt: &decided,
		Decider:   "bot:telegram:123",
		Status:    approvalModel.RequestStatusRejected,
	}
	req.ID = 7

	assert.Contains(t, approverApprovedText(req, now.Add(30*time.Minute)), "已通过 #7")
	assert.Contains(t, approverRejectedText(req, time.Minute), "非工作时间")
	assert.Contains(t, approverRequestGoneText(9), "#9")

	already := approverAlreadyDecidedText(req)
	assert.Contains(t, already, "已拒绝")
	assert.Contains(t, already, "bot:telegram:123")
}

// TestApplicantReplyTexts 申请人回执文案
func TestApplicantReplyTexts(t *testing.T) {
	now := time.Now()
	req := &approvalModel.ApprovalRequest{Username: "alice", ExpireAt: now.Add(2 * time.Minute)}
	req.ID = 3

	assert.Contains(t, applicantSubmittedText(req, now), "#3")
	assert.Contains(t, applicantApprovedText(req, now.Add(30*time.Minute)), "已通过")
	assert.Contains(t, applicantRejectedText(req, time.Minute), "已被拒绝")
}
