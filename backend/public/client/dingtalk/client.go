package dingtalk

import (
	"micro-net-hub/config"
	"micro-net-hub/public/common"

	"github.com/zhaoyunxing92/dingtalk/v2"
)

func InitDingTalkClient() *dingtalk.DingTalk {
	dingTalk, err := dingtalk.NewClient(config.Conf.DingTalk.AppKey, config.Conf.DingTalk.AppSecret)
	if err != nil {
		common.Log.Error("init dingding client failed, err:%v\n", err)
	}
	return dingTalk
}
