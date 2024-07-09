package main

import (
	"context"
	"os"
	"time"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/dnssrv"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/global/setup"
	"micro-net-hub/internal/module/operationlog"
	operationLogModel "micro-net-hub/internal/module/operationlog/model"
	"micro-net-hub/internal/radiussrv"
	"micro-net-hub/internal/server"

	"golang.org/x/sync/errgroup"
)

var (
	g      errgroup.Group
	ctx    context.Context
	cancel func()
)

func main() {

	// 加载配置文件到全局配置结构体
	config.InitConfig()

	// 初始化日志
	setup.InitLogger()

	// global.Log.Debugf("%+v", config.Conf)
	// 初始化数据库(mysql)
	setup.InitDB()

	// setup LdapPool
	setup.InitLdapPool()

	// 初始化casbin策略管理器
	setup.InitCasbinEnforcer()

	// 初始化Validator数据校验
	setup.InitValidate()

	// 初始化mysql数据
	setup.InitData()

	// 启动定时任务
	setup.InitCron()

	// 操作日志中间件处理日志时没有将日志发送到rabbitmq或者kafka中, 而是发送到了channel中
	// 这里开启3个goroutine处理channel将日志记录到数据库
	for i := 0; i < 3; i++ {
		go operationLogModel.SaveOperationLogChannel(operationlog.OperationLogChan)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		return server.Run()
	})

	g.Go(func() error {
		return radiussrv.Run()
	})

	g.Go(func() error {
		return dnssrv.RunTcp()
	})

	g.Go(func() error {
		return dnssrv.RunUdp()
	})

	if err := g.Wait(); err != nil {
		global.Log.Fatal("Server forced to shutdown:", err)
		os.Exit(1)
	}
}
