package radiusappr

import (
	"testing"
	"time"

	"micro-net-hub/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustLocation 加载测试用固定时区
func mustLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	require.NoError(t, err)
	return loc
}

// TestInApprovalWindow 覆盖未开启/全天/日内/跨天/星期限定/配置错误等判定分支
func TestInApprovalWindow(t *testing.T) {
	loc := mustLocation(t, "Asia/Shanghai")
	// 2026-01-01 为周四, 01-03 为周六, 01-04 为周日, 01-05 为周一, 01-06 为周二
	tests := []struct {
		name     string
		approval *config.RadiusApproval
		now      time.Time
		want     bool
		wantErr  bool
	}{
		{
			name:     "审批未开启时恒不判定",
			approval: &config.RadiusApproval{Enable: false, TimeWindows: []config.ApprovalTimeWindow{{Start: "00:00", End: "23:59"}}},
			now:      time.Date(2026, 1, 1, 12, 0, 0, 0, loc),
			want:     false,
		},
		{
			name:     "审批配置为空",
			approval: nil,
			now:      time.Date(2026, 1, 1, 12, 0, 0, 0, loc),
			want:     false,
		},
		{
			name:     "未配置时间窗口视为全天需要审批",
			approval: &config.RadiusApproval{Enable: true},
			now:      time.Date(2026, 1, 1, 3, 0, 0, 0, loc),
			want:     true,
		},
		{
			name: "日内窗口内命中",
			approval: &config.RadiusApproval{
				Enable:      true,
				Timezone:    "Asia/Shanghai",
				TimeWindows: []config.ApprovalTimeWindow{{Start: "09:00", End: "18:00"}},
			},
			now:  time.Date(2026, 1, 1, 10, 0, 0, 0, loc),
			want: true,
		},
		{
			name: "日内窗口外不命中",
			approval: &config.RadiusApproval{
				Enable:      true,
				Timezone:    "Asia/Shanghai",
				TimeWindows: []config.ApprovalTimeWindow{{Start: "09:00", End: "18:00"}},
			},
			now:  time.Date(2026, 1, 1, 20, 0, 0, 0, loc),
			want: false,
		},
		{
			name: "窗口起点包含、终点不包含",
			approval: &config.RadiusApproval{
				Enable:      true,
				Timezone:    "Asia/Shanghai",
				TimeWindows: []config.ApprovalTimeWindow{{Start: "09:00", End: "18:00"}},
			},
			now:  time.Date(2026, 1, 1, 9, 0, 0, 0, loc),
			want: true,
		},
		{
			name: "跨天窗口在凌晨命中",
			approval: &config.RadiusApproval{
				Enable:      true,
				Timezone:    "Asia/Shanghai",
				TimeWindows: []config.ApprovalTimeWindow{{Start: "22:00", End: "06:00"}},
			},
			now:  time.Date(2026, 1, 2, 1, 0, 0, 0, loc),
			want: true,
		},
		{
			name: "跨天窗口在白天不命中",
			approval: &config.RadiusApproval{
				Enable:      true,
				Timezone:    "Asia/Shanghai",
				TimeWindows: []config.ApprovalTimeWindow{{Start: "22:00", End: "06:00"}},
			},
			now:  time.Date(2026, 1, 2, 12, 0, 0, 0, loc),
			want: false,
		},
		{
			name: "星期限定窗口在指定星期命中",
			approval: &config.RadiusApproval{
				Enable:      true,
				Timezone:    "Asia/Shanghai",
				TimeWindows: []config.ApprovalTimeWindow{{Days: []string{"mon"}, Start: "09:00", End: "18:00"}},
			},
			now:  time.Date(2026, 1, 5, 10, 0, 0, 0, loc),
			want: true,
		},
		{
			name: "星期限定窗口在其它星期不命中",
			approval: &config.RadiusApproval{
				Enable:      true,
				Timezone:    "Asia/Shanghai",
				TimeWindows: []config.ApprovalTimeWindow{{Days: []string{"mon"}, Start: "09:00", End: "18:00"}},
			},
			now:  time.Date(2026, 1, 6, 10, 0, 0, 0, loc),
			want: false,
		},
		{
			name: "跨天窗口的星期按锚定日判定",
			approval: &config.RadiusApproval{
				Enable:      true,
				Timezone:    "Asia/Shanghai",
				TimeWindows: []config.ApprovalTimeWindow{{Days: []string{"sat"}, Start: "22:00", End: "06:00"}},
			},
			// 周日凌晨 2 点落在周六 22:00 起始的跨天窗口内
			now:  time.Date(2026, 1, 4, 2, 0, 0, 0, loc),
			want: true,
		},
		{
			name: "多个窗口任意命中即需审批",
			approval: &config.RadiusApproval{
				Enable:   true,
				Timezone: "Asia/Shanghai",
				TimeWindows: []config.ApprovalTimeWindow{
					{Start: "00:00", End: "06:00"},
					{Days: []string{"sat", "sun"}, Start: "09:00", End: "18:00"},
				},
			},
			now:  time.Date(2026, 1, 3, 10, 0, 0, 0, loc),
			want: true,
		},
		{
			name: "时间格式非法",
			approval: &config.RadiusApproval{
				Enable:      true,
				TimeWindows: []config.ApprovalTimeWindow{{Start: "25:00", End: "06:00"}},
			},
			now:     time.Date(2026, 1, 1, 1, 0, 0, 0, loc),
			wantErr: true,
		},
		{
			name: "星期取值非法",
			approval: &config.RadiusApproval{
				Enable:      true,
				TimeWindows: []config.ApprovalTimeWindow{{Days: []string{"monday"}, Start: "09:00", End: "18:00"}},
			},
			now:     time.Date(2026, 1, 1, 10, 0, 0, 0, loc),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InApprovalWindow(tt.now, tt.approval)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestLoadLocation 时区解析: 留空回落到本地时区, 非法值同样回落而不是报错
func TestLoadLocation(t *testing.T) {
	assert.Same(t, time.Local, loadLocation(""))
	assert.Same(t, time.Local, loadLocation("Not/AZone"))

	loc := loadLocation("Asia/Shanghai")
	require.NotNil(t, loc)
	assert.Equal(t, "Asia/Shanghai", loc.String())
}
