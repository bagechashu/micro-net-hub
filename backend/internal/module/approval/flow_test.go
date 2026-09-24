package approval

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	accountModel "micro-net-hub/internal/module/account/model"
	approvalModel "micro-net-hub/internal/module/approval/model"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// setupApprovalEnv 准备内存数据库与审批配置.
//
// 每个用例使用独立的共享内存库(名称带纳秒后缀), 避免用例之间互相污染; 全局的
// global.DB / global.Log / config.Conf.Approval 在清理时恢复, 保证用例可重复执行.
func setupApprovalEnv(t *testing.T, approval *config.ApprovalConfig) {
	t.Helper()

	dsn := fmt.Sprintf("file:approval-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	// 单连接: 内存库的连接切换会导致数据不可见
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, db.AutoMigrate(&approvalModel.ApprovalRequest{}, &approvalModel.ApprovalGrant{}))

	previousDB, previousLog := global.DB, global.Log
	global.DB = db
	global.Log = zap.NewNop().Sugar()
	t.Cleanup(func() {
		global.DB = previousDB
		global.Log = previousLog
	})

	previousApproval := config.Conf.Approval
	config.Conf.Approval = approval
	t.Cleanup(func() { config.Conf.Approval = previousApproval })

	// 频控与冷却缓存是包级状态, 逐用例重置避免相互影响
	rejectCooldownCache.Flush()
	commandLimiter.Store(newRateLimiter(5, 10*time.Second))
}

// TestCreateOrReuseRequest 新建与复用待审批申请单
func TestCreateOrReuseRequest(t *testing.T) {
	setupApprovalEnv(t, &config.ApprovalConfig{Enable: true, PendingTTLSeconds: 120})
	user := &accountModel.User{Username: "alice", Nickname: "艾丽斯"}

	req, created, err := CreateOrReuseRequest(user, Meta{MetaKeySourceAddr: "10.0.0.1", MetaKeySourceID: "ocserv-1"})
	require.NoError(t, err)
	require.True(t, created)
	assert.Equal(t, approvalModel.RequestStatusPending, req.Status)
	assert.Equal(t, "艾丽斯", req.Nickname)
	assert.Equal(t, "10.0.0.1", metaGet(req, MetaKeySourceAddr))

	reused, created, err := CreateOrReuseRequest(user, Meta{MetaKeySourceAddr: "10.0.0.2"})
	require.NoError(t, err)
	assert.False(t, created, "同一用户名的重复请求应复用待审批申请单")
	assert.Equal(t, req.ID, reused.ID)
	assert.Equal(t, "10.0.0.1", metaGet(reused, MetaKeySourceAddr), "复用时不覆盖原始来源信息")
}

// TestCreateOrReuseRequest_RejectsNilUser 非法入参必须报错而不是落库
func TestCreateOrReuseRequest_RejectsNilUser(t *testing.T) {
	setupApprovalEnv(t, &config.ApprovalConfig{Enable: true})

	_, _, err := CreateOrReuseRequest(nil, Meta{})
	require.Error(t, err)
}

// TestCreateOrReuseRequest_MaxPending 全局待审批上限生效
func TestCreateOrReuseRequest_MaxPending(t *testing.T) {
	setupApprovalEnv(t, &config.ApprovalConfig{Enable: true, PendingTTLSeconds: 120, MaxPendingGlobal: 1})

	_, _, err := CreateOrReuseRequest(&accountModel.User{Username: "alice"}, Meta{})
	require.NoError(t, err)

	_, _, err = CreateOrReuseRequest(&accountModel.User{Username: "bob"}, Meta{})
	require.ErrorIs(t, err, ErrTooManyPending)
}

// TestCreateOrReuseRequest_AfterExpire 已过期的申请单不可复用, 应新建
func TestCreateOrReuseRequest_AfterExpire(t *testing.T) {
	setupApprovalEnv(t, &config.ApprovalConfig{Enable: true, PendingTTLSeconds: 120})
	user := &accountModel.User{Username: "alice"}

	first, _, err := CreateOrReuseRequest(user, Meta{})
	require.NoError(t, err)

	require.NoError(t, global.DB.Model(&approvalModel.ApprovalRequest{}).Where("id = ?", first.ID).
		Update("expire_at", time.Now().Add(-time.Minute)).Error)

	second, created, err := CreateOrReuseRequest(user, Meta{})
	require.NoError(t, err)
	require.True(t, created)
	assert.NotEqual(t, first.ID, second.ID)

	old, err := approvalModel.FindApprovalRequestByID(first.ID)
	require.NoError(t, err)
	require.NotNil(t, old)
	assert.Equal(t, approvalModel.RequestStatusExpired, old.Status)
}

// TestApproveIssuesGrantAndIsIdempotent 审批通过签发凭证, 重复审批为幂等操作
func TestApproveIssuesGrantAndIsIdempotent(t *testing.T) {
	setupApprovalEnv(t, &config.ApprovalConfig{Enable: true, PendingTTLSeconds: 120, GrantTTLMinutes: 30})

	req, _, err := CreateOrReuseRequest(&accountModel.User{Username: "alice"}, Meta{})
	require.NoError(t, err)

	decision, err := Approve(req.ID, "bot:telegram:1", approvalModel.ChannelBotTelegram, "同意")
	require.NoError(t, err)
	require.True(t, decision.Applied)
	require.NotNil(t, decision.Grant)
	assert.Equal(t, approvalModel.RequestStatusApproved, decision.Request.Status)

	grant, err := FindActiveGrant("alice")
	require.NoError(t, err)
	require.NotNil(t, grant)
	assert.Equal(t, req.ID, grant.RequestID)

	again, err := Approve(req.ID, "bot:telegram:2", approvalModel.ChannelBotTelegram, "")
	require.NoError(t, err)
	assert.False(t, again.Applied, "第二次审批应为幂等重复操作")
	assert.Equal(t, approvalModel.RequestStatusApproved, again.Request.Status)
	assert.Equal(t, "bot:telegram:1", again.Request.Decider, "首次审批人不应被覆盖")
}

// TestRejectSetsCooldown 审批拒绝后进入冷却期
func TestRejectSetsCooldown(t *testing.T) {
	setupApprovalEnv(t, &config.ApprovalConfig{Enable: true, PendingTTLSeconds: 120, RejectCooldownSeconds: 60})

	req, _, err := CreateOrReuseRequest(&accountModel.User{Username: "alice"}, Meta{})
	require.NoError(t, err)

	decision, err := Reject(req.ID, "bot:telegram:1", approvalModel.ChannelBotTelegram, "非工作时间")
	require.NoError(t, err)
	assert.True(t, decision.Applied)
	assert.Nil(t, decision.Grant, "拒绝不应签发凭证")
	assert.Equal(t, "非工作时间", decision.Request.Reason)

	assert.Positive(t, RejectCooldownLeft("alice").Seconds())
	assert.Zero(t, RejectCooldownLeft("bob"), "未申请过的用户不应处于冷却期")

	// 冷却期内不会再次产生新申请单
	allowed, err := Gate(context.Background(), &accountModel.User{Username: "alice"}, Meta{})
	assert.False(t, allowed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "已被拒绝")
}

// TestGate_AllowsWithActiveGrant 命中放行凭证时直接放行并记录使用次数
func TestGate_AllowsWithActiveGrant(t *testing.T) {
	setupApprovalEnv(t, &config.ApprovalConfig{Enable: true, WaitSeconds: 1, GrantTTLMinutes: 30})

	grant, err := GrantAccess("alice", "bot:telegram:1", approvalModel.ChannelBotTelegram, time.Minute)
	require.NoError(t, err)
	require.NotNil(t, grant)

	allowed, err := Gate(context.Background(), &accountModel.User{Username: "alice"}, Meta{})
	require.NoError(t, err)
	assert.True(t, allowed)

	latest, err := approvalModel.FindActiveGrant("alice", time.Now())
	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, 1, latest.UsedCount, "放行后应记录一次使用")
}

// TestGate_WaitsForApproval 门禁在等待窗口内被审批通过后应立即放行
func TestGate_WaitsForApproval(t *testing.T) {
	setupApprovalEnv(t, &config.ApprovalConfig{Enable: true, WaitSeconds: 5, PendingTTLSeconds: 120})

	result := make(chan error, 1)
	go func() {
		allowed, err := Gate(context.Background(), &accountModel.User{Username: "alice"}, Meta{})
		switch {
		case err != nil:
			result <- err
		case !allowed:
			result <- errors.New("门禁未放行")
		default:
			result <- nil
		}
	}()

	var req *approvalModel.ApprovalRequest
	require.Eventually(t, func() bool {
		req, _ = approvalModel.FindPendingRequestByUsername("alice")
		return req != nil
	}, 3*time.Second, 20*time.Millisecond, "门禁应创建待审批申请单")

	decision, err := Approve(req.ID, "bot:telegram:1", approvalModel.ChannelBotTelegram, "同意")
	require.NoError(t, err)
	require.True(t, decision.Applied)

	select {
	case err := <-result:
		require.NoError(t, err, "审批通过后门禁应放行")
	case <-time.After(3 * time.Second):
		t.Fatal("门禁未在审批通过后返回")
	}
}

// TestGate_TimesOutAndKeepsRequest 等待超时不放行, 但申请单保留供审批人稍后处理
func TestGate_TimesOutAndKeepsRequest(t *testing.T) {
	setupApprovalEnv(t, &config.ApprovalConfig{Enable: true, WaitSeconds: 1, PendingTTLSeconds: 120})

	allowed, err := Gate(context.Background(), &accountModel.User{Username: "alice"}, Meta{})
	require.False(t, allowed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "人工审批")

	req, err := approvalModel.FindPendingRequestByUsername("alice")
	require.NoError(t, err)
	require.NotNil(t, req, "超时后申请单应保留")

	// 审批人稍后处理, 用户凭凭证重连即可放行
	decision, err := Approve(req.ID, "bot:telegram:1", approvalModel.ChannelBotTelegram, "补批")
	require.NoError(t, err)
	require.True(t, decision.Applied)
}

// TestGate_SkipsOutsideScope 不在审批范围内的用户直接放行
func TestGate_SkipsOutsideScope(t *testing.T) {
	setupApprovalEnv(t, &config.ApprovalConfig{
		Enable:      true,
		WaitSeconds: 1,
		Scope:       config.ApprovalScope{ExcludeUsers: []string{"alice"}},
	})

	allowed, err := Gate(context.Background(), &accountModel.User{Username: "alice"}, Meta{})
	require.NoError(t, err)
	assert.True(t, allowed)

	req, err := approvalModel.FindPendingRequestByUsername("alice")
	require.NoError(t, err)
	assert.Nil(t, req, "不在审批范围内不应创建申请单")
}

// TestGate_SkipsOutsideTimeWindow 未命中时间窗口时直接放行
func TestGate_SkipsOutsideTimeWindow(t *testing.T) {
	loc := mustLocation(t, "Asia/Shanghai")
	now := time.Now().In(loc)

	// 构造一个当前时刻必然不命中的窗口: 起点为 2 小时后, 终点为 3 小时后
	setupApprovalEnv(t, &config.ApprovalConfig{
		Enable:      true,
		Timezone:    loc.String(),
		WaitSeconds: 1,
		TimeWindows: []config.ApprovalTimeWindow{{
			Start: now.Add(2 * time.Hour).Format("15:04"),
			End:   now.Add(3 * time.Hour).Format("15:04"),
		}},
	})

	allowed, err := Gate(context.Background(), &accountModel.User{Username: "alice"}, Meta{})
	require.NoError(t, err)
	assert.True(t, allowed)

	req, err := approvalModel.FindPendingRequestByUsername("alice")
	require.NoError(t, err)
	assert.Nil(t, req, "未命中时间窗口不应创建申请单")
}

// TestGrantAccessHandlesPendingRequest 应急放行应同时处理待审批申请单
func TestGrantAccessHandlesPendingRequest(t *testing.T) {
	setupApprovalEnv(t, &config.ApprovalConfig{Enable: true, PendingTTLSeconds: 120, GrantTTLMinutes: 30})

	req, _, err := CreateOrReuseRequest(&accountModel.User{Username: "alice"}, Meta{})
	require.NoError(t, err)

	grant, err := GrantAccess("alice", "bot:telegram:1", approvalModel.ChannelBotTelegram, 10*time.Minute)
	require.NoError(t, err)
	require.NotNil(t, grant)
	assert.WithinDuration(t, time.Now().Add(10*time.Minute), grant.ExpireAt, time.Minute)

	handled, err := approvalModel.FindApprovalRequestByID(req.ID)
	require.NoError(t, err)
	require.NotNil(t, handled)
	assert.Equal(t, approvalModel.RequestStatusApproved, handled.Status)

	pending, err := approvalModel.FindPendingRequestByUsername("alice")
	require.NoError(t, err)
	assert.Nil(t, pending, "已放行用户的待审批申请单不应残留")
}

// TestCleanupExpiresOverdueRequests 清理任务把超时申请单置为已过期
func TestCleanupExpiresOverdueRequests(t *testing.T) {
	setupApprovalEnv(t, &config.ApprovalConfig{Enable: true, PendingTTLSeconds: 120})

	req, _, err := CreateOrReuseRequest(&accountModel.User{Username: "alice"}, Meta{})
	require.NoError(t, err)

	require.NoError(t, global.DB.Model(&approvalModel.ApprovalRequest{}).Where("id = ?", req.ID).
		Update("expire_at", time.Now().Add(-time.Minute)).Error)

	Cleanup()

	latest, err := approvalModel.FindApprovalRequestByID(req.ID)
	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, approvalModel.RequestStatusExpired, latest.Status)
	assert.Equal(t, approvalModel.ChannelSystem, latest.DeciderChannel)

	pending, err := approvalModel.FindPendingRequestByUsername("alice")
	require.NoError(t, err)
	assert.Nil(t, pending)
}
