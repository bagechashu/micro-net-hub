package user

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

// 注册基础路由
func InitBaseRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	base := r.Group("/base")
	{
		base.GET("/encryptpwd", EncryptPasswd) // 生成加密密码
		// base.GET("/decryptpwd", DecryptPasswd) // 密码解密为明文
		// curl "http://127.0.0.1:9000/api/base/decryptpwd?passwd=$(echo 'D5TYG4U29/iqD+99lkXIxuwP0bHPLz5GvIIVhx9ZHobecad6JDgw2p5EFreRx3UWIwL9WJH1E32yzREdhkWlrZTgQ5GuKTItk34ZPtwnVLc+fgwSJ9OiYvaEzWOQXwMdWdo7sdpE89R4XktHPRXMULiPT3rTC8hT4pGG16RxqRo=' | sed 's/+/%2B/g; s/\//%2F/g; s/=/%3D/g')"

		base.POST("/login", authMiddleware.LoginHandler)   // 登录刷新token无需鉴权
		base.POST("/logout", authMiddleware.LogoutHandler) // 登出刷新token无需鉴权
		base.GET("/refreshToken", authMiddleware.RefreshHandler)
		base.POST("/sendcode", SendCode)   // 给用户邮箱发送验证码
		base.POST("/changePwd", ChangePwd) // 修改用户密码
	}
	return r
}
