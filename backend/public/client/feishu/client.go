package feishu

import (
	"micro-net-hub/config"

	"github.com/chyroc/lark"
)

func InitFeiShuClient() *lark.Lark {
	return lark.New(lark.WithAppCredential(
		config.Conf.FeiShu.AppID,
		config.Conf.FeiShu.AppSecret,
	))
}
