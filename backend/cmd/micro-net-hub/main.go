package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/dnssrv"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/global/setup"
	"micro-net-hub/internal/keyrotate"
	"micro-net-hub/internal/module/operationlog"
	operationLogModel "micro-net-hub/internal/module/operationlog/model"
	"micro-net-hub/internal/radiussrv"
	"micro-net-hub/internal/server"

	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

var (
	g      errgroup.Group
	ctx    context.Context
	cancel func()
)

func main() {
	// 子命令：轮换 RSA 密钥（需先停止服务），见 runRotateRSAKey
	if len(os.Args) > 1 && os.Args[1] == "rotate-rsa-key" {
		os.Exit(runRotateRSAKey(os.Args[2:]))
	}

	// 加载配置文件到全局配置结构体
	config.InitConfig()

	// 初始化日志
	setup.InitLogger()

	// global.Log.Debugf("%+v", config.Conf)
	// 初始化数据库(mysql)
	setup.InitDB()

	// 检查并加载 RSA 密钥：全新部署时自动生成；仍使用仓库公开密钥时告警
	setup.InitRSAKeys()

	// setup LdapPool
	if config.Conf.Ldap.EnableManage {
		setup.InitLdapPool()
	}

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

// runRotateRSAKey 轮换 RSA 密钥：生成新密钥对，并把数据库里的用户口令和 config.yml 里加密的口令改用新密钥加密。
//
//	micro-net-hub rotate-rsa-key [--dry-run]
//
// 执行前请先停止服务（运行中的服务仍持有旧密钥和用户缓存），执行后再启动。
func runRotateRSAKey(args []string) int {
	fs := flag.NewFlagSet("rotate-rsa-key", flag.ContinueOnError)
	dryRun := fs.Bool("dry-run", false, "只检查并统计，不做任何修改")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	config.InitConfig()
	setup.InitLogger()
	setup.ConnectDB() // 只连接，不迁移表结构

	err := keyrotate.Run(global.DB, keyrotate.Options{
		DryRun:     *dryRun,
		ConfigPath: viper.ConfigFileUsed(),
		Out:        os.Stdout,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "轮换失败: %v\n", err)
		return 1
	}
	return 0
}
