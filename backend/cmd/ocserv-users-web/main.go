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
		config    = flag.String("config", "rules.yaml", "配置文件路径 (json|yaml)")
		dbPath    = flag.String("dbpath", "data", "违规日志数据库路径")
		refresh   = flag.Duration("refresh", 30*time.Second, "刷新间隔")
		webListenAddr = flag.String("webaddr", ":8080", "Web服务监听地址")
		timezone  = flag.String("timezone", "UTC", "时区设置用于时间检查 (如: UTC, Asia/Shanghai)")
	)

	flag.Parse()

	// 初始化时区设置（用于 VPN 访问控制的时间检查）
	if err := internal.InitializeTimeLocation(*timezone); err != nil {
		log.Fatalf("[main] 时区初始化失败: %v", err)
	}

	// 初始化违规日志数据库
	if err := internal.InitializeViolationDB(*dbPath); err != nil {
		log.Fatalf("[main] 违规日志数据库初始化失败: %v", err)
	}
	defer internal.CloseViolationDB()

	// 加载配置（包含防火墙规则和 VPN 访问控制规则）
	cfg, err := internal.LoadConfig(*config)
	if err != nil {
		log.Fatalf("[main] 配置加载失败: %v", err)
	}

	// 初始化 nftables
	log.Println("[nft] nftables 初始化")
	publicRules := internal.GetPublicRules(cfg)
	inputChainRules := internal.GetInputChainRules(cfg)
	srcIpSets, inputChainIpSetRules := internal.GetInputChainIpSetRules(cfg)
	if err := internal.InitNftables(publicRules, inputChainRules, inputChainIpSetRules, srcIpSets); err != nil {
		log.Fatalf("[main] nftables 初始化失败: %v", err)
	}

	// 启动后初始化所有用户的规则
	userRules := internal.GetUserRulesMapping(cfg)
	internal.UpdateUserRules(userRules)
	if err := internal.UpdateNftablesRulesWithSessions(userRules); err != nil {
		log.Printf("[nft] 更新规则失败: %v", err)
	}

	// 启动VPN访问控制器
	if len(cfg.VpnAccessRules) > 0 {
		internal.UpdateVpnAccessRules(cfg.VpnAccessRules)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go internal.RunVpnAccessTimeEnforcer(ctx, *refresh)
	} else {
		log.Println("[main] 未提供 vpn access 相关配置，跳过访问控制")
	}

	// 启动 nftables 管理器
	// ctx, cancel := context.WithCancel(context.Background()) // 创建可取消的上下文
	// defer cancel()
	// go internal.RunNftablesManager(ctx, *refresh)

	// 初始化全局配置管理器
	web.GlobalConfigManager = internal.NewConfigManager(*config)

	// 启动 WEB 服务
	go web.RunWebServer(*webListenAddr, cfg)

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("[main] 收到退出信号，正在关闭...")
}
