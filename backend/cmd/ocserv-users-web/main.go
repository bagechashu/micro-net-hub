// env GOOS=linux GOARCH=amd64 go build -o ocserv-users-linux-amd64
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"ocserv-users/internal"
	"ocserv-users/web"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		config  = flag.String("config", "rules.yaml", "配置文件路径 (json|yaml)")
		refresh             = flag.Duration("refresh", 30*time.Second, "刷新间隔")
		webListenAddr       = flag.String("webaddr", ":8080", "Web服务监听地址")
	)

	flag.Parse()

	// 加载防火墙配置
	nftRulesCfg, err := internal.LoadConfig(*config)
	if err != nil {
		log.Fatalf("[main] 配置加载失败: %v", err)
	}

	// 先在启动时加载并校验配置文件，确保配置有效
	internal.Global_VpnAccessRules, err = internal.LoadVpnAccessConfig(*config)
	if err != nil {
		log.Fatalf("[main] vpn access 配置加载失败: %v", err)
	}
	if internal.Global_VpnAccessRules == nil {
		log.Println("[main] 未提供 vpn access 相关配置 ，跳过访问控制")
	}

	// 初始化 nftables
	log.Println("[nft] nftables 初始化")
	publicRules := internal.GetPublicRules(nftRulesCfg)
	inputChainRules := internal.GetInputChainRules(nftRulesCfg)
	srcIpSets, inputChainIpSetRules := internal.GetInputChainIpSetRules(nftRulesCfg)
	if err := internal.InitNftables(publicRules, inputChainRules, inputChainIpSetRules, srcIpSets); err != nil {
		log.Fatalf("[main] nftables 初始化失败: %v", err)
	}

	// 启动后初始化所有用户的规则
	internal.Global_UsersRules = internal.GetUserRulesMapping(nftRulesCfg)
	if err := internal.UpdateNftablesRulesWithSessions(internal.Global_UsersRules); err != nil {
		log.Printf("[nft] 更新规则失败: %v", err)
	}

	// 启动VPN访问控制器
	if internal.Global_VpnAccessRules != nil {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go internal.RunVpnAccessTimeEnforcer(ctx, *refresh)
	}

	// 启动 nftables 管理器
	// ctx, cancel := context.WithCancel(context.Background()) // 创建可取消的上下文
	// defer cancel()
	// go internal.RunNftablesManager(ctx, *refresh)

	// 启动 WEB 服务
	go web.RunWebServer(*webListenAddr)

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("[main] 收到退出信号，正在关闭...")
}
