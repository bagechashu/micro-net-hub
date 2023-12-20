package main

import (
	"fmt"

	"github.com/eryajf/go-ldap-admin/config"
	"github.com/eryajf/go-ldap-admin/public/tools"
)

func main() {
	// 加载配置文件到全局配置结构体
	config.InitConfig()
	fmt.Printf("admin passwd encrypted string: %s", tools.NewGenPasswd(config.Conf.Ldap.AdminPass))
}
