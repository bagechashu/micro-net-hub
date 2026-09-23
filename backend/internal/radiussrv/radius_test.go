package radiussrv

import (
	"context"
	"testing"
	"time"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	accountModel "micro-net-hub/internal/module/account/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// withApproval 在测试期间替换全局审批配置与日志, 结束后恢复
func withApproval(t *testing.T, approval *config.RadiusApproval) {
	t.Helper()

	previousRadius, previousLog := config.Conf.Radius, global.Log
	config.Conf.Radius = &config.Radius{Approval: approval}
	global.Log = zap.NewNop().Sugar()
	t.Cleanup(func() {
		config.Conf.Radius = previousRadius
		global.Log = previousLog
	})
}

// TestCheckApproval_Disabled 审批未开启时门禁直接放行, 且不触碰数据库
func TestCheckApproval_Disabled(t *testing.T) {
	withApproval(t, nil)

	require.NoError(t, checkApproval(context.Background(), &accountModel.User{Username: "alice"}, AuthMeta{}))
}

// TestCheckApproval_OutsideTimeWindow 不在审批时间窗口内时直接放行
func TestCheckApproval_OutsideTimeWindow(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	now := time.Now().In(loc)

	// 起点为 2 小时后、终点为 3 小时后: 当前时刻必然不命中
	withApproval(t, &config.RadiusApproval{
		Enable:   true,
		Timezone: loc.String(),
		TimeWindows: []config.ApprovalTimeWindow{{
			Start: now.Add(2 * time.Hour).Format("15:04"),
			End:   now.Add(3 * time.Hour).Format("15:04"),
		}},
	})

	require.NoError(t, checkApproval(context.Background(), &accountModel.User{Username: "alice"}, AuthMeta{RemoteAddr: "10.0.0.1"}))
}

// TestCheckApproval_InvalidTimeWindow 时间窗口配置非法时保守拒绝, 避免静默放行
func TestCheckApproval_InvalidTimeWindow(t *testing.T) {
	withApproval(t, &config.RadiusApproval{
		Enable: true,
		TimeWindows: []config.ApprovalTimeWindow{{
			Start: "25:99",
			End:   "26:99",
		}},
	})

	err := checkApproval(context.Background(), &accountModel.User{Username: "alice"}, AuthMeta{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "时间窗口")
}
