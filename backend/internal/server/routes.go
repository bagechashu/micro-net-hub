package server

import (
	"fmt"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/module/account/auth"
	"micro-net-hub/internal/module/account/group"
	"micro-net-hub/internal/module/account/menu"
	"micro-net-hub/internal/module/account/role"
	"micro-net-hub/internal/module/account/user"
	"micro-net-hub/internal/module/apimgr"
	"micro-net-hub/internal/module/dashboard"
	"micro-net-hub/internal/module/dns"
	fieldrelation "micro-net-hub/internal/module/goldap/field_relation"
	"micro-net-hub/internal/module/goldap/sync"
	"micro-net-hub/internal/module/noticeboard"
	"micro-net-hub/internal/module/operationlog"
	"micro-net-hub/internal/module/sitenav"
	"micro-net-hub/ui"

	"github.com/gin-gonic/gin"
)

// 初始化
func InitRoutes(r *gin.Engine) {
	// 初始化JWT认证中间件
	authMiddleware, err := auth.InitAuthMiddleware()
	if err != nil {
		global.Log.Panicf("初始化JWT中间件失败：%v", err)
		panic(fmt.Sprintf("初始化JWT中间件失败：%v", err))
	}

	// 路由分组
	apiGroup := r.Group("/" + config.Conf.System.UrlPathPrefix)

	// 注册路由
	// 注册嵌入式 UI 路由, 不需要jwt认证中间件,不需要casbin中间件
	{
		ui.InitUiRoutes(r)

	}

	// Account module routes
	{
		dashboard.InitDashboardRoutes(apiGroup, authMiddleware) // 注册dashboard路由, 不需要jwt认证中间件,不需要casbin中间件
		user.InitBaseRoutes(apiGroup, authMiddleware)           // 注册基础路由, 不需要jwt认证中间件,不需要casbin中间件
		user.InitUserRoutes(apiGroup, authMiddleware)           // 注册用户路由, jwt认证中间件,casbin鉴权中间件
		group.InitGroupRoutes(apiGroup, authMiddleware)         // 注册分组路由, jwt认证中间件,casbin鉴权中间件
		role.InitRoleRoutes(apiGroup, authMiddleware)           // 注册角色路由, jwt认证中间件,casbin鉴权中间件
		menu.InitMenuRoutes(apiGroup, authMiddleware)           // 注册菜单路由, jwt认证中间件,casbin鉴权中间件

	}

	// ApiMgr module routes
	{
		apimgr.InitApiRoutes(apiGroup, authMiddleware) // 注册接口路由, jwt认证中间件,casbin鉴权中间件

	}

	// OperationLog module routes
	{
		operationlog.InitOperationLogRoutes(apiGroup, authMiddleware) // 注册操作日志路由, jwt认证中间件,casbin鉴权中间件

	}

	// Goldap module routes
	{
		sync.InitGoldapSyncRoutes(apiGroup, authMiddleware)                   // 注册goldap同步路由, jwt认证中间件,casbin鉴权中间件
		fieldrelation.InitGoldapFieldRelationRoutes(apiGroup, authMiddleware) // 注册字段关联路由, jwt认证中间件,casbin鉴权中间件
	}

	// SiteNav module routes
	{
		sitenav.InitSiteNavRoutes(apiGroup, authMiddleware)
	}

	// Dns Manager module routes
	{
		dns.InitDnsMgrRoutes(apiGroup, authMiddleware)
	}

	// Notice Manager module routes
	{
		noticeboard.InitNoticeMgrRoutes(apiGroup, authMiddleware)
	}

	global.Log.Info("初始化路由完成！")
}
