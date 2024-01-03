package test

import (
	"fmt"
	"testing"

	"micro-net-hub/internal/global/setup"
	userModel "micro-net-hub/internal/module/user/model"
	"micro-net-hub/internal/server/config"
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

	var u userModel.UserService
	filter := tools.H{
		"id": "111",
	}

	if u.Exist(filter) {
		fmt.Println("用户名已存在")
	} else {
		fmt.Println("用户名不存在")
	}
}
