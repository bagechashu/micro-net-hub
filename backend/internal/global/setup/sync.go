package setup

import (
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/module/goldap/sync"
	"micro-net-hub/internal/module/goldap/usermgr"

	"github.com/robfig/cron/v3"
)

func InitCron() {
	c := cron.New(cron.WithSeconds())

	if config.Conf.DingTalk != nil && config.Conf.DingTalk.Flag != "" && config.Conf.Sync.EnableSync {
		ding := usermgr.NewDingTalk()

		//启动定时任务
		_, err := c.AddFunc(config.Conf.Sync.DeptSyncTime, func() {
			ding.SyncDepts()
		})
		if err != nil {
			global.Log.Errorf("启动同步部门的定时任务失败: %v", err)
		}
		//每天凌晨1点执行一次
		_, err = c.AddFunc(config.Conf.Sync.UserSyncTime, func() {
			ding.SyncUsers()
		})
		if err != nil {
			global.Log.Errorf("启动同步用户的定时任务失败: %v", err)
		}
	}

	if config.Conf.WeCom != nil && config.Conf.WeCom.Flag != "" && config.Conf.Sync.EnableSync {
		wechat := usermgr.NewWeChat()

		_, err := c.AddFunc(config.Conf.Sync.DeptSyncTime, func() {
			wechat.SyncDepts()
		})
		if err != nil {
			global.Log.Errorf("启动同步部门的定时任务失败: %v", err)
		}
		//每天凌晨1点执行一次
		_, err = c.AddFunc(config.Conf.Sync.UserSyncTime, func() {
			wechat.SyncUsers()
		})
		if err != nil {
			global.Log.Errorf("启动同步用户的定时任务失败: %v", err)
		}
	}

	if config.Conf.FeiShu != nil && config.Conf.FeiShu.Flag != "" && config.Conf.Sync.EnableSync {
		feishu := usermgr.NewFeiShu()

		_, err := c.AddFunc(config.Conf.Sync.DeptSyncTime, func() {
			feishu.SyncDepts()
		})
		if err != nil {
			global.Log.Errorf("启动同步部门的定时任务失败: %v", err)
		}
		//每天凌晨1点执行一次
		_, err = c.AddFunc(config.Conf.Sync.UserSyncTime, func() {
			feishu.SyncUsers()
		})
		if err != nil {
			global.Log.Errorf("启动同步用户的定时任务失败: %v", err)
		}
	}

	// 自动检索未同步数据
	global.Log.Infof("启动同步ldap数据的定时任务: %s", config.Conf.Sync.LdapSyncTime)
	_, err := c.AddFunc(config.Conf.Sync.LdapSyncTime, func() {
		_ = sync.SearchGroupDiff()
		_ = sync.SearchUserDiff()
	})
	if err != nil {
		global.Log.Errorf("启动同步任务状态检查任务失败: %v", err)
	}
	c.Start()
	global.Log.Info("初始化定时任务完成")
}
