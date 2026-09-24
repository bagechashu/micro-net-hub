// Package approval 实现登录认证的人工审批: 时间窗口判定、申请单状态机、
// 放行凭证与 Bot(Telegram) 审批交互.
package approval

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	accountModel "micro-net-hub/internal/module/account/model"
	approvalModel "micro-net-hub/internal/module/approval/model"

	"github.com/patrickmn/go-cache"
)

// Meta 认证请求携带的上下文信息, 由各认证入口(如 LDAP/HTTP 等)自行构造的
// 键值对, 以 JSON 落库留痕, 便于不同认证场景携带不同字段.
//
// 约定 key(可选, 用于审批人通知展示):
//
//	MetaKeySourceAddr 来源地址
//	MetaKeySourceID   来源设备标识
//
// 其余 key 仅作审计留存, 不在通知中展示.
type Meta map[string]string

// Meta 约定 key
const (
	MetaKeySourceAddr = "sourceAddr" // 来源地址(展示)
	MetaKeySourceID   = "sourceId"   // 来源设备标识(展示)
	MetaKeySourceIP   = "sourceIp"   // 来源设备地址(仅留痕)
)

// Decision 一次审批决定的执行结果
type Decision struct {
	Request *approvalModel.ApprovalRequest // 决定后的申请单
	Grant   *approvalModel.ApprovalGrant   // 审批通过时签发的放行凭证
	Applied bool                           // 本次调用是否真正做出了决定(否则为幂等重复操作)
}

// ErrTooManyPending 全局待审批申请数超限
var ErrTooManyPending = errors.New("待审批的申请过多, 请稍后再试")

// createRequestMu 串行化申请单创建, 消除「查无 pending 再创建」在并发下各自建一条
// 的竞态. 该锁仅覆盖单进程, 多实例部署需配合数据库唯一约束兜底.
var createRequestMu sync.Mutex

// rejectCooldownCache 记录用户的拒绝冷却时间
var rejectCooldownCache = cache.New(10*time.Minute, 20*time.Minute)

// timeouts 审批相关的超时与有效期参数
type timeouts struct {
	wait           time.Duration // 认证侧同步等待时长
	pendingTTL     time.Duration // 申请单有效期
	noticeInterval time.Duration // 通知审批人的最小间隔
	rejectCooldown time.Duration // 拒绝后的申请冷却时长
	grantTTL       time.Duration // 放行凭证有效期
	grantMaxUses   int           // 放行凭证最大使用次数, 0 表示不限次
	maxPending     int           // 全局待审批申请上限
}

// configApproval 返回当前生效的审批配置, 未配置时返回 nil
func configApproval() *config.ApprovalConfig {
	if config.Conf == nil {
		return nil
	}
	return config.Conf.Approval
}

// Enabled 判断人工审批是否已开启
func Enabled() bool {
	approval := configApproval()
	return approval != nil && approval.Enable
}

// approvalTimeouts 返回审批相关参数, 配置缺省项回落到内置默认值
func approvalTimeouts() timeouts {
	out := timeouts{
		wait:           6 * time.Second,
		pendingTTL:     120 * time.Second,
		noticeInterval: 15 * time.Second,
		rejectCooldown: 60 * time.Second,
		grantTTL:       30 * time.Minute,
		grantMaxUses:   0,
		maxPending:     200,
	}

	approval := configApproval()
	if approval == nil {
		return out
	}
	if approval.WaitSeconds > 0 {
		out.wait = time.Duration(approval.WaitSeconds) * time.Second
	}
	if approval.PendingTTLSeconds > 0 {
		out.pendingTTL = time.Duration(approval.PendingTTLSeconds) * time.Second
	}
	if approval.NoticeIntervalSeconds > 0 {
		out.noticeInterval = time.Duration(approval.NoticeIntervalSeconds) * time.Second
	}
	if approval.RejectCooldownSeconds > 0 {
		out.rejectCooldown = time.Duration(approval.RejectCooldownSeconds) * time.Second
	}
	if approval.GrantTTLMinutes > 0 {
		out.grantTTL = time.Duration(approval.GrantTTLMinutes) * time.Minute
	}
	out.grantMaxUses = approval.GrantMaxUses
	if approval.MaxPendingGlobal > 0 {
		out.maxPending = approval.MaxPendingGlobal
	}
	return out
}

// CreateOrReuseRequest 创建或复用某用户的待审批申请单.
//
// created 为 true 表示本次新建了申请单(应通知审批人), false 表示复用了已有申请单,
// 从而避免客户端重试导致审批人被反复打扰; 是否重新通知由 ShouldNotice 决定.
func CreateOrReuseRequest(user *accountModel.User, meta Meta) (req *approvalModel.ApprovalRequest, created bool, err error) {
	if user == nil {
		return nil, false, errors.New("创建审批申请失败: 用户信息为空")
	}

	createRequestMu.Lock()
	defer createRequestMu.Unlock()

	t := approvalTimeouts()
	now := time.Now()

	existing, err := approvalModel.FindPendingRequestByUsername(user.Username)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		if existing.ExpireAt.After(now) {
			return existing, false, nil
		}
		// 复用一条已过期的申请单没有意义, 先置为过期再新建
		if _, err := approvalModel.DecideRequest(existing.ID, approvalModel.RequestStatusExpired,
			"system", approvalModel.ChannelSystem, "申请单已过期", now); err != nil {
			return nil, false, err
		}
	}

	pendingCount, err := approvalModel.CountPendingRequests()
	if err != nil {
		return nil, false, err
	}
	if t.maxPending > 0 && pendingCount >= int64(t.maxPending) {
		return nil, false, ErrTooManyPending
	}

	if meta == nil {
		meta = Meta{}
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return nil, false, fmt.Errorf("序列化认证上下文失败: %w", err)
	}

	req = &approvalModel.ApprovalRequest{
		Username:   user.Username,
		Nickname:   user.Nickname,
		Meta:       string(metaJSON),
		Status:     approvalModel.RequestStatusPending,
		ExpireAt:   now.Add(t.pendingTTL),
		PendingKey: approvalModel.PendingKeyFor(user.Username),
	}
	if err := approvalModel.CreateApprovalRequest(req); err != nil {
		// 数据库唯一约束兜底: 多实例部署下并发创建同一用户的申请单时, 后到者会被
		// pending_key 唯一索引拒绝; 此时重查一次复用已存在的那条, 而非直接报错.
		if existing, findErr := approvalModel.FindPendingRequestByUsername(user.Username); findErr == nil && existing != nil {
			return existing, false, nil
		}
		return nil, false, err
	}
	return req, true, nil
}

// MarkNoticed 记录申请单的本次通知, 用于通知限频
func MarkNoticed(req *approvalModel.ApprovalRequest, now time.Time) {
	if req == nil {
		return
	}
	req.NoticeCount++
	req.LastNoticeAt = &now
	if err := approvalModel.UpdateNoticeInfo(req.ID, req.NoticeCount, now); err != nil {
		global.Log.Errorf("更新审批申请通知状态失败, id=%d: %v", req.ID, err)
	}
}

// ShouldNotice 判断该申请单此刻是否应通知审批人(受通知间隔约束)
func ShouldNotice(req *approvalModel.ApprovalRequest, now time.Time) bool {
	if req == nil {
		return false
	}
	if req.LastNoticeAt == nil {
		return true
	}
	return now.Sub(*req.LastNoticeAt) >= approvalTimeouts().noticeInterval
}

// rejectCooldownKey 拒绝冷却缓存的键
func rejectCooldownKey(username string) string {
	return fmt.Sprintf("reject:%s", username)
}

// Approve 审批通过: 原子地决定申请单并签发放行凭证
func Approve(requestID uint, decider, channel, reason string) (*Decision, error) {
	now := time.Now()
	t := approvalTimeouts()

	winner, err := approvalModel.DecideRequest(requestID, approvalModel.RequestStatusApproved,
		decider, channel, reason, now)
	if err != nil {
		return nil, err
	}
	if !winner {
		// 已被其他审批人处理, 返回最新状态用于幂等回执
		latest, err := approvalModel.FindApprovalRequestByID(requestID)
		if err != nil {
			return nil, err
		}
		return &Decision{Request: latest, Applied: false}, nil
	}

	req, err := approvalModel.FindApprovalRequestByID(requestID)
	if err != nil {
		return nil, err
	}

	grant := &approvalModel.ApprovalGrant{
		Username:  req.Username,
		RequestID: req.ID,
		ExpireAt:  now.Add(t.grantTTL),
		MaxUses:   t.grantMaxUses,
		GrantedBy: decider,
		Channel:   channel,
	}
	if err := approvalModel.CreateApprovalGrant(grant); err != nil {
		// 凭证签发失败不回滚审批结论: 申请单已通过, 审批人可让用户重连或重新审批
		global.Log.Errorf("签发放行凭证失败, username=%s, requestID=%d: %v", req.Username, req.ID, err)
		grant = nil
	}

	global.Log.Infof("登录审批通过: requestID=%d, username=%s, decider=%s, channel=%s",
		req.ID, req.Username, decider, channel)
	wakeWaiters(req.ID, approvalModel.RequestStatusApproved)
	return &Decision{Request: req, Grant: grant, Applied: true}, nil
}

// Reject 审批拒绝: 原子地决定申请单并记录冷却时间
func Reject(requestID uint, decider, channel, reason string) (*Decision, error) {
	now := time.Now()
	t := approvalTimeouts()

	winner, err := approvalModel.DecideRequest(requestID, approvalModel.RequestStatusRejected,
		decider, channel, reason, now)
	if err != nil {
		return nil, err
	}
	if !winner {
		latest, err := approvalModel.FindApprovalRequestByID(requestID)
		if err != nil {
			return nil, err
		}
		return &Decision{Request: latest, Applied: false}, nil
	}

	req, err := approvalModel.FindApprovalRequestByID(requestID)
	if err != nil {
		return nil, err
	}

	if t.rejectCooldown > 0 {
		rejectCooldownCache.Set(rejectCooldownKey(req.Username), now, t.rejectCooldown)
	}

	global.Log.Infof("登录审批拒绝: requestID=%d, username=%s, decider=%s, channel=%s, reason=%s",
		req.ID, req.Username, decider, channel, reason)
	wakeWaiters(req.ID, approvalModel.RequestStatusRejected)
	return &Decision{Request: req, Applied: true}, nil
}

// RejectCooldownLeft 返回用户剩余的拒绝冷却时长, 0 表示不在冷却期
func RejectCooldownLeft(username string) time.Duration {
	value, found := rejectCooldownCache.Get(rejectCooldownKey(username))
	if !found {
		return 0
	}
	left := approvalTimeouts().rejectCooldown - time.Since(value.(time.Time))
	if left < 0 {
		return 0
	}
	return left
}

// FindActiveGrant 查询用户当前有效的放行凭证
func FindActiveGrant(username string) (*approvalModel.ApprovalGrant, error) {
	return approvalModel.FindActiveGrant(username, time.Now())
}

// GrantAccess 应急放行: 直接为用户签发放行凭证, 并同时处理其待审批申请单
func GrantAccess(username, decider, channel string, ttl time.Duration) (*approvalModel.ApprovalGrant, error) {
	now := time.Now()
	t := approvalTimeouts()
	if ttl <= 0 {
		ttl = t.grantTTL
	}

	grant := &approvalModel.ApprovalGrant{
		Username:  username,
		ExpireAt:  now.Add(ttl),
		MaxUses:   t.grantMaxUses,
		GrantedBy: decider,
		Channel:   channel,
	}
	if err := approvalModel.CreateApprovalGrant(grant); err != nil {
		return nil, err
	}

	// 已放行的用户不应再挂着待审批申请单
	pending, err := approvalModel.FindPendingRequestByUsername(username)
	if err == nil && pending != nil {
		if _, err := approvalModel.DecideRequest(pending.ID, approvalModel.RequestStatusApproved,
			decider, channel, "管理员应急放行", now); err != nil {
			global.Log.Errorf("应急放行时处理待审批申请失败, requestID=%d: %v", pending.ID, err)
		} else {
			wakeWaiters(pending.ID, approvalModel.RequestStatusApproved)
		}
	}

	global.Log.Infof("应急放行: username=%s, decider=%s, channel=%s, expireAt=%s",
		username, decider, channel, formatTime(grant.ExpireAt))
	return grant, nil
}

// ListPending 获取待审批申请列表
func ListPending(limit int) ([]*approvalModel.ApprovalRequest, error) {
	if limit <= 0 {
		limit = 10
	}
	return approvalModel.ListPendingRequests(limit)
}

// Cleanup 清理过期数据: 超时的待审批申请单置为过期, 并删除超过保留期的历史数据
func Cleanup() {
	now := time.Now()

	expired, err := approvalModel.ExpirePendingRequestsBefore(now)
	if err != nil {
		global.Log.Errorf("清理过期审批申请失败: %v", err)
		return
	}

	// 保留 30 天历史, 便于回溯审计
	retainBefore := now.AddDate(0, 0, -30)
	reqs, err := approvalModel.CleanupRequestsBefore(retainBefore)
	if err != nil {
		global.Log.Errorf("清理历史审批申请失败: %v", err)
	}
	grants, err := approvalModel.CleanupGrantsBefore(retainBefore)
	if err != nil {
		global.Log.Errorf("清理历史放行凭证失败: %v", err)
	}

	global.Log.Infof("清理审批数据完成: 置为过期的申请 %d 条, 删除申请 %d 条, 删除凭证 %d 条",
		expired, reqs, grants)
}
