// env GOOS=linux GOARCH=amd64 go build -o ocserv-users-linux-amd64
package main

import (
	"flag"
	"log"
	"ocserv-users/internal"
	"ocserv-users/web"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		configPath = flag.String("config", "rules.json", "配置文件路径")
		// refresh       = flag.Duration("refresh", 30*time.Second, "刷新间隔")
		webListenAddr = flag.String("webaddr", ":8080", "Web服务监听地址")
	)

	flag.Parse()

	// 加载防火墙配置
	cfg, err := internal.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("[main] 配置加载失败: %v", err)
	}
	// log.Printf("[main] 配置加载成功: %+v", cfg)

	// 初始化 nftables
	log.Println("[nft] nftables 初始化")
	publicRules := internal.GetPublicRules(cfg)
	inputChainRules := internal.GetInputChainRules(cfg)
	if err := internal.InitNftables(publicRules, inputChainRules); err != nil {
		log.Fatalf("[main] nftables 初始化失败: %v", err)
	}

	// 启动后初始化所有用户的规则
	internal.GlobalUserRules = internal.GetUserRulesMapping(cfg)
	if err := internal.UpdateNftablesRulesWithSessions(internal.GlobalUserRules); err != nil {
		log.Printf("[nft] 更新规则失败: %v", err)
	}

	// 启动 nftables 管理器
	// 创建可取消的上下文
	// ctx, cancel := context.WithCancel(context.Background())
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
