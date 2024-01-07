package setup

import (
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/module/goldap/usermgr"
	"micro-net-hub/internal/module/user"

	"github.com/robfig/cron/v3"
)

func InitCron() {
	c := cron.New(cron.WithSeconds())

	ding := usermgr.NewDingTalk()
	wechat := usermgr.NewWeChat()
	feishu := usermgr.NewFeiShu()
	if config.Conf.DingTalk.EnableSync {
		//启动定时任务
		_, err := c.AddFunc(config.Conf.DingTalk.DeptSyncTime, func() {
			ding.SyncDepts(nil, nil)
		})
		if err != nil {
			global.Log.Errorf("启动同步部门的定时任务失败: %v", err)
		}
		//每天凌晨1点执行一次
		_, err = c.AddFunc(config.Conf.DingTalk.UserSyncTime, func() {
			ding.SyncUsers(nil, nil)
		})
		if err != nil {
			global.Log.Errorf("启动同步用户的定时任务失败: %v", err)
		}
	}
	if config.Conf.WeCom.EnableSync {
		_, err := c.AddFunc(config.Conf.WeCom.DeptSyncTime, func() {
			wechat.SyncDepts(nil, nil)
		})
		if err != nil {
			global.Log.Errorf("启动同步部门的定时任务失败: %v", err)
		}
		//每天凌晨1点执行一次
		_, err = c.AddFunc(config.Conf.WeCom.UserSyncTime, func() {
			wechat.SyncUsers(nil, nil)
		})
		if err != nil {
			global.Log.Errorf("启动同步用户的定时任务失败: %v", err)
		}
	}
	if config.Conf.FeiShu.EnableSync {
		_, err := c.AddFunc(config.Conf.FeiShu.DeptSyncTime, func() {
			feishu.SyncDepts(nil, nil)
		})
		if err != nil {
			global.Log.Errorf("启动同步部门的定时任务失败: %v", err)
		}
		//每天凌晨1点执行一次
		_, err = c.AddFunc(config.Conf.FeiShu.UserSyncTime, func() {
			feishu.SyncUsers(nil, nil)
		})
		if err != nil {
			global.Log.Errorf("启动同步用户的定时任务失败: %v", err)
		}
	}

	// 自动检索未同步数据
	_, err := c.AddFunc("0 */2 * * * *", func() {
		// 开发调试时调整为10秒执行一次
		// _, err := c.AddFunc("*/10 * * * * *", func() {
		_ = user.SearchGroupDiff()
		_ = user.SearchUserDiff()
	})
	if err != nil {
		global.Log.Errorf("启动同步任务状态检查任务失败: %v", err)
	}
	c.Start()
}
