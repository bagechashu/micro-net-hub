package operationlog

import (
	"fmt"
	"strings"
	"time"

	"micro-net-hub/internal/config"
	accountModel "micro-net-hub/internal/module/account/model"
	apiMgrModel "micro-net-hub/internal/module/apimgr/model"
	"micro-net-hub/internal/module/operationlog/model"

	"github.com/gin-gonic/gin"
)

// 操作日志channel
var OperationLogChan = make(chan *model.OperationLog, 30)

func OperationLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行耗时
		timeCost := endTime.Sub(startTime).Milliseconds()
		// 获取访问路径
		path := strings.TrimPrefix(c.FullPath(), "/"+config.Conf.System.UrlPathPrefix)
		// 请求方式
		method := c.Request.Method

		// TODO: 通过参数配置是否记录 GET 和 嵌入UI 日志
		// 过滤写入数据库的日志
		if method == "GET" {
			return
		} else if strings.HasPrefix(c.FullPath(), "/ui") {
			return
		}

		// 获取当前登录用户
		var username string
		ctxUser, _ := c.Get("user")
		user, ok := ctxUser.(accountModel.User)
		if !ok {
			username = "未登录"
		} else {
			username = user.Username
		}

		// 检查接口并获取其描述
		api := new(apiMgrModel.Api)
		_ = apiMgrModel.Find(map[string]interface{}{"path": path, "method": method}, api)

		operationLog := model.OperationLog{
			Username:   username,
			Ip:         c.ClientIP(),
			IpLocation: "",
			Method:     method,
			Path:       path,
			Remark:     api.Remark,
			Status:     c.Writer.Status(),
			StartTime:  fmt.Sprintf("%v", startTime),
			TimeCost:   timeCost,
			UserAgent:  c.Request.UserAgent(),
		}

		// 最好是将日志发送到rabbitmq或者kafka中
		// 这里是发送到channel中，开启3个goroutine处理
		OperationLogChan <- &operationLog
	}
}
