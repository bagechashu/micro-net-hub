package setup

import (
	"context"
	"strings"

	"micro-net-hub/internal/bot"
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/radiusappr"
)

// InitBot 初始化 Bot 实例管理器并按配置启动实例, 返回停止函数供进程退出时调用.
//
// 单个实例启动失败只记录日志, 不影响进程启动: Bot 是审批等能力的通道, 不应成为
// 整个服务可用性的前置条件. 使用独立的 context(而非 main 里的超时 ctx), 否则
// 实例会在超时后被立刻取消.
func InitBot() context.CancelFunc {
	if config.Conf.Bot == nil || !config.Conf.Bot.Enable {
		global.Log.Info("Bot 接入未启用, 跳过 Bot 实例初始化")
		return func() {}
	}

	specs := botInstanceSpecs()
	if len(specs) == 0 {
		global.Log.Info("Bot 接入已启用, 但没有可用的实例配置")
		return func() {}
	}

	manager := bot.NewManager(bot.DefaultRegistry(), radiusappr.BotHandler)
	global.BotManager = manager

	// 审批模块通过列表器获取全部运行实例, 从而不直接依赖全局管理器
	radiusappr.SetProviderLister(manager.Providers)

	ctx, cancel := context.WithCancel(context.Background())
	if err := manager.Start(ctx, specs); err != nil {
		global.Log.Errorf("部分 Bot 实例启动失败: %v", err)
	}
	global.Log.Infof("Bot 实例初始化完成, 运行中: %v", manager.Statuses())

	return func() {
		cancel()
		manager.StopAll()
		global.Log.Info("Bot 实例已全部停止")
	}
}

// botInstanceSpecs 把配置转换为 Bot 实例运行规格, 跳过禁用与缺少 id 的实例
func botInstanceSpecs() []bot.InstanceSpec {
	instances := config.Conf.Bot.Instances
	specs := make([]bot.InstanceSpec, 0, len(instances))

	for _, instance := range instances {
		if !instance.Enable {
			global.Log.Infof("Bot 实例 [%s] 已禁用, 跳过", instance.ID)
			continue
		}

		id := strings.TrimSpace(instance.ID)
		if id == "" {
			global.Log.Warnf("Bot 实例缺少 id 字段, 跳过: type=%s, name=%s", instance.Type, instance.Name)
			continue
		}

		specs = append(specs, bot.InstanceSpec{
			ID:     id,
			Type:   strings.TrimSpace(instance.Type),
			Name:   instance.Name,
			Config: instance.Config,
		})
	}
	return specs
}
