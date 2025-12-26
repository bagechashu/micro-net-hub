package ui

import (
	"embed"
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

//go:embed index.html favicon.ico static assets
var Static embed.FS

// cacheMiddleware 为静态文件添加缓存头
func cacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求路径是否为静态资源
		if strings.HasPrefix(c.Request.URL.Path, "/ui/static/") {
			// 设置缓存头 - 7天的缓存时间
			c.Header("Cache-Control", "public, max-age=604800")
			// 设置Expires头
			expires := time.Now().AddDate(0, 0, 7) // 7天后过期
			c.Header("Expires", expires.Format(time.RFC1123))
		} else if strings.HasPrefix(c.Request.URL.Path, "/ui/assets/") {
			// 对于assets目录也设置长时间缓存 - 30天
			c.Header("Cache-Control", "public, max-age=2592000")
			expires := time.Now().AddDate(0, 0, 30)
			c.Header("Expires", expires.Format(time.RFC1123))
		} else {
			// 对于HTML文件等动态内容设置不缓存
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		c.Next()
	}
}

func InitUiRoutes(r *gin.Engine) gin.IRoutes {
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "ui/")
	})

	ui := r.Group("/ui", cacheMiddleware())

	// because I use embedded file system, so use hardcode path
	if config.Conf.System.Mode == "release" {
		ui.StaticFS("/", http.FS(Static))
		global.Log.Info("release Mode, and Webui was embedded static file")
	} else {
		ui.StaticFS("/", http.Dir("./ui/"))
	}
	return r
}
