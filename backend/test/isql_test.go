package test

import (
	"fmt"
	"testing"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global/setup"
	accountModel "micro-net-hub/internal/module/account/model"
	"micro-net-hub/internal/tools"
)

func InitConfig() {
	// 加载配置文件到全局配置结构体
	config.InitConfig()

	// 初始化日志
	setup.InitLogger()

	// 初始化数据库(mysql)
	setup.InitDB()

	// 初始化casbin策略管理器
	setup.InitCasbinEnforcer()

	// 初始化Validator数据校验
	setup.InitValidate()
}

func TestUserExist(t *testing.T) {
	InitConfig()

	var u accountModel.User
	filter := tools.H{
		"id": "111",
	}

	if u.Exist(filter) {
		fmt.Println("用户名已存在")
	} else {
		fmt.Println("用户名不存在")
	}
}
