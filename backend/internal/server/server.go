package server

import (
	"fmt"
	"net/http"
	"time"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/server/routes"
)

type GinServer struct {
	Server *http.Server
}

func NewGinServer() *http.Server {
	// 注册所有路由
	r := routes.InitRoutes()

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
