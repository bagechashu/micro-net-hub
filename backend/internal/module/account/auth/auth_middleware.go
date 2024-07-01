package auth

import (
	"fmt"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/module/account/model"
	"micro-net-hub/internal/server/helper"
	"micro-net-hub/internal/tools"

	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

// 初始化jwt中间件
func InitAuthMiddleware() (*jwt.GinJWTMiddleware, error) {
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:           config.Conf.Jwt.Realm,                                      // jwt标识
		Key:             []byte(config.Conf.Jwt.Key),                                // 服务端密钥
		Timeout:         time.Minute * time.Duration(config.Conf.Jwt.TimeoutMin),    // token过期时间
		MaxRefresh:      time.Minute * time.Duration(config.Conf.Jwt.MaxRefreshMin), // token最大刷新时间(RefreshToken过期时间=Timeout+MaxRefresh)
		PayloadFunc:     payloadFunc,                                                // 有效载荷处理
		IdentityHandler: identityHandler,                                            // 解析Claims
		Authenticator:   authenticator,                                              // 校验token的正确性, 处理登录逻辑
		Authorizator:    authorizator,                                               // 用户登录校验成功处理
		Unauthorized:    unauthorized,                                               // 用户登录校验失败处理
		LoginResponse:   loginResponse,                                              // 登录成功后的响应
		LogoutResponse:  logoutResponse,                                             // 登出后的响应
		RefreshResponse: refreshResponse,                                            // 刷新token后的响应
		TokenLookup:     "header: Authorization, query: token, cookie: jwt",         // 自动在这几个地方寻找请求中的token
		TokenHeadName:   "Bearer",                                                   // header名称
		TimeFunc:        time.Now().UTC,
	})
	return authMiddleware, err
}

const customClaimsName string = "ukey"

// 有效载荷处理
func payloadFunc(data interface{}) jwt.MapClaims {
	if v, ok := data.(map[string]interface{}); ok {
		var user model.User
		// 将用户json转为结构体
		tools.JsonI2Struct(v[customClaimsName], &user)
		return jwt.MapClaims{
			jwt.IdentityKey:  user.ID,
			customClaimsName: v[customClaimsName],
		}
	}
	return jwt.MapClaims{}
}

// 解析Claims
func identityHandler(c *gin.Context) interface{} {
	claims := jwt.ExtractClaims(c)
	// 此处返回值类型map[string]interface{}与payloadFunc和authorizator的data类型必须一致, 否则会导致授权失败还不容易找到原因
	return map[string]interface{}{
		"IdentityKey":    claims[jwt.IdentityKey],
		customClaimsName: claims[customClaimsName],
	}
}

// RegisterAndLoginReq 用户登录结构体
type RegisterAndLoginReq struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

// 登录逻辑, 校验token的正确性
func authenticator(c *gin.Context) (interface{}, error) {
	var req RegisterAndLoginReq
	// 请求json绑定
	if err := c.ShouldBind(&req); err != nil {
		return "", err
	}

	// 密码通过RSA解密
	decodeData, err := tools.RSADecrypt([]byte(req.Password), config.Conf.System.RSAPrivateBytes)
	if err != nil {
		return nil, err
	}

	u := &model.User{
		Username: req.Username,
		Password: string(decodeData),
	}

	// 密码校验
	userLogined, err := u.Login()
	if err != nil {
		return nil, err
	}
	// 将用户以json格式写入, payloadFunc/authorizator会使用到
	return map[string]interface{}{
		customClaimsName: tools.Struct2Json(
			struct {
				ID       uint
				Username string
			}{
				ID:       userLogined.ID,
				Username: userLogined.Username,
			},
		),
	}, nil
}

// 用户登录校验成功处理
func authorizator(data interface{}, c *gin.Context) bool {
	if v, ok := data.(map[string]interface{}); ok {
		userClaims := v[customClaimsName].(string)
		var user model.User
		// 将用户json转为结构体
		tools.Json2Struct(userClaims, &user)

		// 将用户保存到context, api调用时取数据方便
		c.Set(customClaimsName, user)
		return true
	}
	return false
}

// 用户登录校验失败处理
func unauthorized(c *gin.Context, code int, message string) {
	global.Log.Debugf("JWT认证失败, 错误码: %d, 错误信息: %s", code, message)
	helper.Response(c, code, code, nil, fmt.Sprintf("JWT认证失败, 错误码: %d, 错误信息: %s", code, message))
}

// 登录成功后的响应
func loginResponse(c *gin.Context, code int, token string, expires time.Time) {
	helper.Response(c, code, code,
		gin.H{
			"token":   token,
			"expires": expires.Format("2006-01-02 15:04:05"),
		},
		"登录成功")
}

// 登出后的响应
func logoutResponse(c *gin.Context, code int) {
	helper.SuccessWithMessage(c, nil, "退出成功")
}

// 刷新token后的响应
func refreshResponse(c *gin.Context, code int, token string, expires time.Time) {
	helper.Response(c, code, code,
		gin.H{
			"token":   token,
			"expires": expires,
		},
		"刷新token成功")
}
