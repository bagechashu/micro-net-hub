package role

import (
	"strings"
	"sync"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/module/account/auth"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

var checkLock sync.Mutex

// Casbin中间件, 基于RBAC的权限访问控制模型
func CasbinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctxuser, err := auth.GetCtxLoginUser(c)
		if err != nil {
			helper.Response(c, 401, 401, nil, "用户未登录")
			c.Abort()
			return
		}
		if ctxuser.Status != 1 {
			helper.Response(c, 401, 401, nil, "当前用户已被禁用")
			c.Abort()
			return
		}
		// 获得用户的全部角色
		roles := ctxuser.Roles
		// 获得用户全部未被禁用的角色的Keyword
		var subs []string
		for _, role := range roles {
			if role.Status == 1 {
				subs = append(subs, role.Keyword)
			}
		}
		// 获得请求路径URL
		obj := strings.TrimPrefix(c.FullPath(), "/"+config.Conf.System.UrlPathPrefix)
		// 获取请求方式
		act := c.Request.Method
		isPass := check(subs, obj, act)
		if !isPass {
			helper.Response(c, 401, 401, nil, "您当前的角色没有权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

func check(subs []string, obj string, act string) bool {
	// 同一时间只允许一个请求执行校验, 否则可能会校验失败
	checkLock.Lock()
	defer checkLock.Unlock()
	isPass := false
	for _, sub := range subs {
		pass, _ := global.CasbinEnforcer.Enforce(sub, obj, act)
		if pass {
			isPass = true
			break
		}
	}
	return isPass
}
