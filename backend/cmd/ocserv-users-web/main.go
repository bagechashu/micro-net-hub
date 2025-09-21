// main.go
package main

import (
	"context"
	"flag"
	"log"
	"ocserv-users/internal"
	"ocserv-users/web"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var (
		configPath    = flag.String("config", "rules.json", "配置文件路径")
		refresh       = flag.Duration("refresh", 30*time.Second, "刷新间隔")
		webListenAddr = flag.String("webaddr", ":8080", "Web服务监听地址")
	)

	flag.Parse()

	cfg, err := internal.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("[main] 配置加载失败: %v", err)
	}

	// 初始化 nftables
	if err := internal.InitNftables(cfg); err != nil {
		log.Fatalf("[main] nftables 初始化失败: %v", err)
	}

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动服务
	go internal.RunNftablesManager(ctx, cfg, *refresh)
	go web.RunWebServer(*webListenAddr)

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("[main] 收到退出信号，正在关闭...")
}
