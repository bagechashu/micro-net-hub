package config

import (
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSampleConfigs_ParseApproval 校验随仓库提供的各份配置样例都能解析出审批与 Bot 配置.
//
// 配置样例是运维部署的直接依据: 一旦字段名调整而样例未同步, 线上就会出现
// "配置看着写了但不生效"的静默故障, 这里应当先失败.
func TestSampleConfigs_ParseApproval(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "backend/config.yml", path: "../../config.yml"},
		{name: "docker-compose/simple", path: "../../../docs/docker-compose/simple/config/app/config.yml"},
		{name: "docker-compose/all-in-one", path: "../../../docs/docker-compose/all-in-one/config/app/config.yml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := loadConfig(t, tt.path)

			require.NotNil(t, conf.Radius, "radius 段缺失")
			require.NotNil(t, conf.Approval, "approval 段缺失")

			approval := conf.Approval
			assert.Equal(t, "Asia/Shanghai", approval.Timezone)
			assert.Positive(t, approval.WaitSeconds)
			assert.Positive(t, approval.PendingTTLSeconds)
			assert.Positive(t, approval.NoticeIntervalSeconds)
			assert.Positive(t, approval.RejectCooldownSeconds)
			assert.Positive(t, approval.GrantTTLMinutes)
			assert.Positive(t, approval.MaxPendingGlobal)
			assert.NotEmpty(t, approval.CleanupCron)

			require.Len(t, approval.TimeWindows, 2, "样例应包含工作日与周末两个窗口")
			assert.Equal(t, []string{"mon", "tue", "wed", "thu", "fri"}, approval.TimeWindows[0].Days)
			assert.NotEmpty(t, approval.TimeWindows[0].Start)
			assert.NotEmpty(t, approval.TimeWindows[0].End)
			assert.NotEmpty(t, approval.TimeWindows[1].Start)
			assert.NotEmpty(t, approval.TimeWindows[1].End)

			assert.Empty(t, approval.Scope.Users, "样例默认不限制审批范围")

			assert.True(t, approval.Bot.Enable)
			assert.True(t, approval.Bot.PrivateOnly, "私聊限制必须默认开启")
			assert.True(t, approval.Bot.ApplicantNotify)
			assert.Positive(t, approval.Bot.MaxCommandsPer10s)
			assert.NotEmpty(t, approval.Bot.InstanceIDs, "审批 Bot 实例 ID 不能为空")

			// 审批引用的实例必须在 bot.instances 中真实存在
			require.NotNil(t, conf.Bot, "bot 段缺失")
			ids := make([]string, 0, len(conf.Bot.Instances))
			for _, instance := range conf.Bot.Instances {
				ids = append(ids, instance.ID)
				assert.NotEmpty(t, instance.Type, "Bot 实例缺少 type")
			}
			for _, id := range approval.Bot.InstanceIDs {
				assert.Contains(t, ids, id, "审批配置引用了不存在的 Bot 实例")
			}
		})
	}
}

// TestDefaults_Approval 默认值键必须能映射进配置结构体(键名写错会在此失败)
func TestDefaults_Approval(t *testing.T) {
	setViperDefaults()

	out := new(config)
	require.NoError(t, viper.Unmarshal(out))
	require.NotNil(t, out.Radius, "radius 默认值未生效")
	require.NotNil(t, out.Approval, "approval 默认值未生效")

	approval := out.Approval
	assert.False(t, approval.Enable, "人工审批默认关闭")
	assert.Equal(t, "Asia/Shanghai", approval.Timezone)
	assert.Equal(t, 6, approval.WaitSeconds)
	assert.Equal(t, 120, approval.PendingTTLSeconds)
	assert.Equal(t, 15, approval.NoticeIntervalSeconds)
	assert.Equal(t, 60, approval.RejectCooldownSeconds)
	assert.Equal(t, 30, approval.GrantTTLMinutes)
	assert.Equal(t, 200, approval.MaxPendingGlobal)
	assert.Equal(t, "0 30 4 * * *", approval.CleanupCron)

	assert.True(t, approval.Bot.Enable)
	assert.True(t, approval.Bot.PrivateOnly, "默认仅允许私聊审批")
	assert.True(t, approval.Bot.ApplicantNotify)
	assert.Empty(t, approval.Bot.Notifications, "通知列表默认空")
	assert.Equal(t, 5, approval.Bot.MaxCommandsPer10s)

	require.NotNil(t, out.Bot, "bot 默认值未生效")
	assert.False(t, out.Bot.Enable, "Bot 接入默认关闭")
}

// loadConfig 用独立的 viper 实例解析配置文件, 避免污染全局 viper
func loadConfig(t *testing.T, path string) *config {
	t.Helper()

	abs, err := filepath.Abs(path)
	require.NoError(t, err)

	v := viper.New()
	v.SetConfigFile(abs)
	require.NoError(t, v.ReadInConfig(), "读取配置文件失败: %s", abs)

	out := new(config)
	require.NoError(t, v.Unmarshal(out))
	return out
}
