package approval

import (
	"context"
	"fmt"
	"sync"
	"time"

	"micro-net-hub/internal/global"
	accountModel "micro-net-hub/internal/module/account/model"
	approvalModel "micro-net-hub/internal/module/approval/model"
)

// pollInterval 等待期间轮询数据库的间隔.
//
// 进程内唤醒通道可以做到即时响应, 但多实例部署或申请单由其他进程处理时收不到唤醒,
// 因此保留数据库轮询作为兜底.
const pollInterval = time.Second

// waiterTable 进程内等待表: 申请单 ID -> 等待该决定的通道.
type waiterTable struct {
	sync.Mutex
	chans map[uint][]chan uint8
}

var waiters = waiterTable{chans: make(map[uint][]chan uint8)}

// registerWaiter 注册等待通道
func registerWaiter(requestID uint) chan uint8 {
	ch := make(chan uint8, 1)
	waiters.Lock()
	waiters.chans[requestID] = append(waiters.chans[requestID], ch)
	waiters.Unlock()
	return ch
}

// unregisterWaiter 注销等待通道
func unregisterWaiter(requestID uint, ch chan uint8) {
	waiters.Lock()
	defer waiters.Unlock()

	list := waiters.chans[requestID]
	for i, item := range list {
		if item == ch {
			waiters.chans[requestID] = append(list[:i], list[i+1:]...)
			break
		}
	}
	if len(waiters.chans[requestID]) == 0 {
		delete(waiters.chans, requestID)
	}
}

// wakeWaiters 唤醒等待某申请单决定的等待者(本进程内即时唤醒)
func wakeWaiters(requestID uint, status uint8) {
	waiters.Lock()
	defer waiters.Unlock()

	for _, ch := range waiters.chans[requestID] {
		select {
		case ch <- status:
		default:
		}
	}
	delete(waiters.chans, requestID)
}

// Gate 人工审批门禁, 是认证链路上接入审批的唯一入口.
//
// 返回值语义: allow 为 true 表示放行; 否则 err 描述拒绝原因.
// 调用方必须保证: 审批相关的拒绝(含等待超时)不计入登录失败次数, 否则用户重试几次即被锁定.
func Gate(ctx context.Context, user *accountModel.User, meta Meta) (allow bool, err error) {
	if !Enabled() {
		return true, nil
	}

	approval := configApproval()
	inWindow, err := InApprovalWindow(time.Now(), approval)
	if err != nil {
		// 配置错误时保守拒绝, 并把原因写进日志, 避免静默放行
		global.Log.Errorf("审批时间窗口判定失败: %v", err)
		return false, fmt.Errorf("审批时间窗口配置错误, 拒绝登录")
	}
	if !inWindow {
		return true, nil
	}
	// 角色/分组维度需要预加载关联数据, 仅在配置确实用到了这些维度时才查库
	scopeUser := user
	if len(approval.Scope.Roles) > 0 || len(approval.Scope.Groups) > 0 {
		loaded, loadErr := loadUserScopeInfo(user.Username)
		if loadErr != nil {
			global.Log.Errorf("加载用户审批范围信息失败, username=%s: %v", user.Username, loadErr)
			return false, fmt.Errorf("加载用户审批范围信息失败, 请稍后重试")
		}
		scopeUser = loaded
	}
	if !InApprovalScope(scopeUser, approval.Scope) {
		return true, nil
	}

	// 审批通过后的放行凭证: 有效期内直接放行
	grant, err := FindActiveGrant(user.Username)
	if err != nil {
		global.Log.Errorf("查询放行凭证失败, username=%s: %v", user.Username, err)
		return false, fmt.Errorf("查询放行凭证失败, 请稍后重试")
	}
	if grant != nil {
		// 原子消耗一次使用次数: 在数据库层同时校验有效性(未作废/未过期/未超次数),
		// 避免并发下 MaxUses 被超用.
		consumed, err := approvalModel.ConsumeGrant(grant.ID, time.Now())
		if err != nil {
			// 计数失败不阻塞放行: 凭证本身有效, 只是使用次数未能记录
			global.Log.Errorf("记录放行凭证使用失败, grantID=%d: %v", grant.ID, err)
			return true, nil
		}
		if consumed {
			global.Log.Infof("命中有效放行凭证, 直接放行: username=%s, grantID=%d, expireAt=%s",
				user.Username, grant.ID, formatTime(grant.ExpireAt))
			return true, nil
		}
		// 未消耗成功(已超次数/已过期/已作废), 回落审批流程重新申请
		global.Log.Infof("放行凭证不可用(已耗尽或失效), 回落审批流程: username=%s, grantID=%d",
			user.Username, grant.ID)
	}

	// 拒绝冷却期内不再重复打扰审批人
	if left := RejectCooldownLeft(user.Username); left > 0 {
		return false, fmt.Errorf("登录申请已被拒绝, 请在 %d 秒后重试", int(left.Seconds()))
	}

	req, created, err := CreateOrReuseRequest(user, meta)
	if err != nil {
		global.Log.Errorf("创建审批申请失败, username=%s: %v", user.Username, err)
		return false, fmt.Errorf("创建登录审批申请失败, 请稍后重试")
	}

	// 新建申请或超过通知间隔时通知审批人, 避免客户端重试刷屏
	if created || ShouldNotice(req, time.Now()) {
		NotifyApprovers(ctx, req)
	}

	wait := approvalTimeouts().wait
	status, decided := waitForDecision(ctx, req.ID, wait, req.ExpireAt)
	if !decided {
		global.Log.Infof("等待人工审批超时: requestID=%d, username=%s, wait=%s", req.ID, user.Username, wait)
		return false, fmt.Errorf("登录申请 #%d 已提交人工审批, 审批通过后请在有效期内重新连接", req.ID)
	}

	switch status {
	case approvalModel.RequestStatusApproved:
		global.Log.Infof("人工审批通过, 放行: requestID=%d, username=%s", req.ID, user.Username)
		return true, nil
	case approvalModel.RequestStatusRejected:
		return false, fmt.Errorf("登录申请 #%d 已被拒绝", req.ID)
	default:
		return false, fmt.Errorf("登录申请 #%d 已过期, 请重新连接以发起新申请", req.ID)
	}
}

// waitForDecision 等待人工决定.
//
// 返回 decided 为 false 表示等待超时(审批人未在窗口内响应), 此时申请单依然有效,
// 审批人稍后仍可处理, 用户凭放行凭证重连即可.
func waitForDecision(ctx context.Context, requestID uint, wait time.Duration, expireAt time.Time) (uint8, bool) {
	if wait <= 0 {
		return 0, false
	}

	ch := registerWaiter(requestID)
	defer unregisterWaiter(requestID, ch)

	deadline := time.Now().Add(wait)
	if expireAt.Before(deadline) {
		deadline = expireAt
	}

	// 复用单个 timer 承载轮询间隔, 避免循环内反复 time.After 创建新定时器
	timer := time.NewTimer(pollInterval)
	defer timer.Stop()

	for {
		if status, done := pollDecision(requestID); done {
			return status, true
		}
		if !time.Now().Before(deadline) {
			return 0, false
		}

		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(pollInterval)

		select {
		case <-ctx.Done():
			return 0, false
		case status := <-ch:
			return status, true
		case <-timer.C:
		}
	}
}

// pollDecision 从数据库读取申请单是否已有终态, 未决定时 done 为 false
func pollDecision(requestID uint) (status uint8, done bool) {
	req, err := approvalModel.FindApprovalRequestByID(requestID)
	if err != nil {
		global.Log.Errorf("轮询审批申请状态失败, requestID=%d: %v", requestID, err)
		return 0, false
	}
	if req == nil {
		return 0, false
	}

	switch req.Status {
	case approvalModel.RequestStatusApproved,
		approvalModel.RequestStatusRejected,
		approvalModel.RequestStatusExpired,
		approvalModel.RequestStatusCanceled:
		return req.Status, true
	default:
		return 0, false
	}
}

// loadUserScopeInfo 加载审批范围判定所需的角色与分组关联数据.
//
// 仅在审批范围配置确实用到了 roles / groups 维度时调用, 避免每次审批都多查一次库.
func loadUserScopeInfo(username string) (*accountModel.User, error) {
	user := new(accountModel.User)
	err := global.DB.Where("username = ?", username).
		Preload("Roles").Preload("Groups").First(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}
