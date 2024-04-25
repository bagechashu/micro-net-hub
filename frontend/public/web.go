package ui

import (
	"embed"
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed index.html favicon.ico static assets
var Static embed.FS

func InitUiRoutes(r *gin.Engine) gin.IRoutes {
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "ui/")
	})

	// cause I use embedded file system, so use hardcode path
	if config.Conf.System.Mode == "release" {
		r.StaticFS("/ui", http.FS(Static))
		global.Log.Info("release Mode, and Webui was embedd static file")
	} else {
		r.StaticFS("/ui", http.Dir("./ui/"))
	}
	return r
}
