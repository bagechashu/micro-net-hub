package routes

import (
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/server/config"
	"micro-net-hub/ui"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitUiRoutes(r *gin.Engine) gin.IRoutes {
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "ui/")
	})

	// TODO: Configure the static file path using configuration
	if config.Conf.System.Mode == "release" {
		r.StaticFS("/ui", http.FS(ui.Static))
		global.Log.Info("release Mode, and Webui wse embedd static file")
	} else {
		r.StaticFS("/ui", http.Dir("./ui/"))
	}
	return r
}
