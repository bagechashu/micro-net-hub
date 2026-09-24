package setup

import (
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/module/approval"
	"micro-net-hub/internal/module/goldap/sync"
	"micro-net-hub/internal/module/goldap/usermgr"

	"github.com/robfig/cron/v3"
)

func InitCron() {
	c := cron.New(cron.WithSeconds())
	// 自动检索未同步数据
	global.Log.Infof("定时任务启动: %s", config.Conf.Sync.LdapSyncTime)

	if config.Conf.DingTalk != nil && config.Conf.DingTalk.Flag != "" && config.Conf.Sync.EnableSync {
		ding := usermgr.NewDingTalk()

		//启动定时任务
		_, err := c.AddFunc(config.Conf.Sync.DeptSyncTime, func() {
			_ = ding.SyncDepts()
		})
		if err != nil {
			global.Log.Errorf("启动同步部门的定时任务失败: %v", err)
		}
		//每天凌晨1点执行一次
		_, err = c.AddFunc(config.Conf.Sync.UserSyncTime, func() {
			_ = ding.SyncUsers()
		})
		if err != nil {
			global.Log.Errorf("启动同步用户的定时任务失败: %v", err)
		}
	}

	if config.Conf.WeCom != nil && config.Conf.WeCom.Flag != "" && config.Conf.Sync.EnableSync {
		wechat := usermgr.NewWeChat()

		_, err := c.AddFunc(config.Conf.Sync.DeptSyncTime, func() {
			_ = wechat.SyncDepts()
		})
		if err != nil {
			global.Log.Errorf("启动同步部门的定时任务失败: %v", err)
		}
		//每天凌晨1点执行一次
		_, err = c.AddFunc(config.Conf.Sync.UserSyncTime, func() {
			_ = wechat.SyncUsers()
		})
		if err != nil {
			global.Log.Errorf("启动同步用户的定时任务失败: %v", err)
		}
	}

	if config.Conf.FeiShu != nil && config.Conf.FeiShu.Flag != "" && config.Conf.Sync.EnableSync {
		feishu := usermgr.NewFeiShu()

		_, err := c.AddFunc(config.Conf.Sync.DeptSyncTime, func() {
			_ = feishu.SyncDepts()
		})
		if err != nil {
			global.Log.Errorf("同步部门的定时任务启动失败: %v", err)
		}
		//每天凌晨1点执行一次
		_, err = c.AddFunc(config.Conf.Sync.UserSyncTime, func() {
			_ = feishu.SyncUsers()
		})
		if err != nil {
			global.Log.Errorf("同步用户的定时任务启动失败: %v", err)
		}
	}

	if config.Conf.Ldap.EnableManage {
		_, err := c.AddFunc(config.Conf.Sync.LdapSyncTime, func() {
			_ = sync.SearchGroupDiff()
			_ = sync.SearchUserDiff()
		})
		if err != nil {
			global.Log.Errorf("同步任务状态检查任务启动失败: %v", err)
		}
	}
	// RADIUS 人工审批数据清理: 超时的申请单置为已过期, 并删除超过保留期的历史数据
	if config.Conf.Radius != nil && config.Conf.Radius.Approval != nil && config.Conf.Radius.Approval.Enable {
		cleanupCron := config.Conf.Radius.Approval.CleanupCron
		if cleanupCron == "" {
			cleanupCron = "0 30 4 * * *"
		}
		_, err := c.AddFunc(cleanupCron, approval.Cleanup)
		if err != nil {
			global.Log.Errorf("启动 RADIUS 审批数据清理任务失败: %v", err)
		}
	}

	c.Start()
	global.Log.Info("初始化定时任务完成")
}
