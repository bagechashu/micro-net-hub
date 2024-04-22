package routes

import (
	"fmt"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/module/apimgr"
	"micro-net-hub/internal/module/operationlog"
	"micro-net-hub/internal/module/sitenav"
	"micro-net-hub/internal/server/middleware"

	"github.com/gin-gonic/gin"
)

// 初始化
func InitRoutes(r *gin.Engine) {
	// 初始化JWT认证中间件
	authMiddleware, err := middleware.InitAuth()
	if err != nil {
		global.Log.Panicf("初始化JWT中间件失败：%v", err)
		panic(fmt.Sprintf("初始化JWT中间件失败：%v", err))
	}

	// 路由分组
	apiGroup := r.Group("/" + config.Conf.System.UrlPathPrefix)

	// 注册路由
	InitUiRoutes(r)                                               // 注册基础路由, 不需要jwt认证中间件,不需要casbin中间件
	InitBaseRoutes(apiGroup, authMiddleware)                      // 注册基础路由, 不需要jwt认证中间件,不需要casbin中间件
	InitUserRoutes(apiGroup, authMiddleware)                      // 注册用户路由, jwt认证中间件,casbin鉴权中间件
	InitGroupRoutes(apiGroup, authMiddleware)                     // 注册分组路由, jwt认证中间件,casbin鉴权中间件
	InitRoleRoutes(apiGroup, authMiddleware)                      // 注册角色路由, jwt认证中间件,casbin鉴权中间件
	InitMenuRoutes(apiGroup, authMiddleware)                      // 注册菜单路由, jwt认证中间件,casbin鉴权中间件
	apimgr.InitApiRoutes(apiGroup, authMiddleware)                // 注册接口路由, jwt认证中间件,casbin鉴权中间件
	operationlog.InitOperationLogRoutes(apiGroup, authMiddleware) // 注册操作日志路由, jwt认证中间件,casbin鉴权中间件
	InitFieldRelationRoutes(apiGroup, authMiddleware)             // 注册操作日志路由, jwt认证中间件,casbin鉴权中间件
	sitenav.InitSiteNavRoutes(apiGroup, authMiddleware)

	global.Log.Info("初始化路由完成！")
}
