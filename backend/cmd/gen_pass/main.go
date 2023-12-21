package main

import (
	"fmt"

	"micro-net-hub/config"
	"micro-net-hub/public/tools"
)

func main() {
	// 加载配置文件到全局配置结构体
	config.InitConfig()
	fmt.Printf("admin passwd encrypted string: %s", tools.NewGenPasswd(config.Conf.Ldap.AdminPass))
}
