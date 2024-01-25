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

	r := gin.New()
	r.Use(gin.Recovery())

	// Ginlog Middleware
	r.Use(middleware.GinzapMiddleware(global.BasicLog))

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
		Addr:           fmt.Sprintf("%s:%d", host, port),
		Handler:        r,
		ReadTimeout:    config.Conf.System.ReadTimeout * time.Second,
		WriteTimeout:   config.Conf.System.WriteTimeout * time.Second,
		MaxHeaderBytes: config.Conf.System.MaxHeaderMBytes * 1024 * 1024, // 1024*1024 = MB
	}

	global.Log.Infof("New Gin server on: %s:%d", host, port)
	return srv
}

func Run() error {
	ginServer := NewGinServer()
	return ginServer.ListenAndServe()
}
