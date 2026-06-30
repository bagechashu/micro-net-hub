// env GOOS=linux GOARCH=amd64 go build -o ocserv-users-linux-amd64
package main

import (
	"context"
	"flag"
	"fmt"
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
		config           = flag.String("config", "rules.yaml", "配置文件路径 (json|yaml)")
		dbPath           = flag.String("dbpath", "data", "违规日志数据库路径")
		refresh          = flag.Duration("refresh", 60*time.Second, "刷新间隔")
		webListenAddr    = flag.String("webaddr", ":8080", "Web服务监听地址")
		timezone         = flag.String("timezone", "UTC", "时区设置用于时间检查 (如: UTC, Asia/Shanghai)")
		strictManagement = flag.Bool("strict", true, "是否启用严格管理模式")
		debugMode        = flag.Bool("debug", false, "是否启用调试模式")
		showVersion      = flag.Bool("version", false, "显示版本信息并退出")
	)

	flag.Parse()

	// 如果请求显示版本，则打印并退出
	if *showVersion {
		fmt.Println(internal.GetVersionInfo())
		os.Exit(0)
	}

	// 设置 debug 模式
	internal.SetDebugMode(*debugMode)

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

	// 在正常启动时，也在日志中记录版本信息
	log.Printf("[main] 启动服务 - %s", internal.GetVersionInfo())

	// 初始化违规通知 webhook 配置
	internal.InitializeVpnAccessNoticeConfig(cfg.VpnAccessNotice)

	// 初始化 nftables
	log.Println("[nft] nftables 初始化")
	publicRules, err := internal.GetPublicRules(cfg)
	if err != nil {
		log.Fatalf("[main] 获取公共规则失败: %v", err)
	}
	inputChainRules, err := internal.GetInputChainRules(cfg)
	if err != nil {
		log.Fatalf("[main] 获取输入链规则失败: %v", err)
	}
	srcIpSets, inputChainIpSetRules, err := internal.GetInputChainIpSetRules(cfg)
	if err != nil {
		log.Fatalf("[main] 获取输入链 IP 集规则失败: %v", err)
	}
	if err := internal.InitNftables(publicRules, inputChainRules, inputChainIpSetRules, srcIpSets); err != nil {
		log.Fatalf("[main] nftables 初始化失败: %v", err)
	}

	// 启动后初始化所有用户的规则
	userRules, err := internal.GetUserRulesMapping(cfg)
	if err != nil {
		log.Fatalf("[main] 获取用户规则映射失败: %v", err)
	}
	internal.UpdateUserRules(userRules)
	if err := internal.InitUsersNftablesRules(userRules); err != nil {
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
	internal.GlobalConfigManager = internal.NewConfigManager(*config)

	// 启动 WEB 服务
	go web.RunWebServer(*webListenAddr, cfg, *strictManagement)

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("[main] 收到退出信号，正在关闭...")
}
