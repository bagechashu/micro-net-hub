package server

import (
	"fmt"
	"net/http"
	"time"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/server/routes"

	"micro-net-hub/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

type GinServer struct {
	Server *http.Server
}

func NewGinServer() *http.Server {
	//设置模式
	gin.SetMode(config.Conf.System.Mode)

	// TODO: 使用 global.Log 记录日志
	// 创建带有默认中间件的路由:
	// 日志与恢复中间件
	r := gin.Default()
	// 创建不带中间件的路由:
	// r := gin.New()
	// r.Use(gin.Recovery())

	// 启用限流中间件
	// 默认每50毫秒填充一个令牌，最多填充200个
	fillInterval := time.Duration(config.Conf.RateLimit.FillInterval)
	capacity := config.Conf.RateLimit.Capacity
	r.Use(middleware.RateLimitMiddleware(time.Millisecond*fillInterval, capacity))

	// 启用全局跨域中间件
	r.Use(middleware.CORSMiddleware())

	// 启用操作日志中间件
	r.Use(middleware.OperationLogMiddleware())

	// 注册所有路由
	routes.InitRoutes(r)

	host := config.Conf.System.Host
	port := config.Conf.System.Port

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	global.Log.Infof("New Gin server on: %s:%d", host, port)
	return srv
}

func Run() error {
	ginServer := NewGinServer()
	return ginServer.ListenAndServe()
}
