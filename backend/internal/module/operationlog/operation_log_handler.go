package operationlog

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"micro-net-hub/internal/module/operationlog/model"
	"micro-net-hub/internal/server/helper"
	"micro-net-hub/internal/tools"
)

// OperationLogListReq 操作日志请求结构体
type OperationLogListReq struct {
	Username string `json:"username" form:"username"`
	Ip       string `json:"ip" form:"ip"`
	Path     string `json:"path" form:"path"`
	Status   int    `json:"status" form:"status"`
	PageNum  int    `json:"pageNum" form:"pageNum"`
	PageSize int    `json:"pageSize" form:"pageSize"`
}

type LogListRsp struct {
	Total int64                `json:"total"`
	Logs  []model.OperationLog `json:"logs"`
}

// List 记录列表
func List(c *gin.Context) {
	var req OperationLogListReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取数据列表
	logs, err := model.List(
		&model.OperationLog{
			Username: req.Username,
			Ip:       req.Ip,
			Path:     req.Path,
			Status:   req.Status,
		},
		req.PageNum,
		req.PageSize,
	)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取接口列表失败: %s", err.Error())))
	}

	rets := make([]model.OperationLog, 0)
	for _, log := range logs {
		rets = append(rets, *log)
	}
	count, err := model.Count()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取接口总数失败")))
	}

	helper.Success(c, LogListRsp{
		Total: count,
		Logs:  rets,
	})

}

// OperationLogDeleteReq 批量删除操作日志结构体
type OperationLogDeleteReq struct {
	OperationLogIds []uint `json:"operationLogIds" validate:"required"`
}

// Delete 删除记录
func Delete(c *gin.Context) {
	var req OperationLogDeleteReq

	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	for _, id := range req.OperationLogIds {
		filter := tools.H{"id": int(id)}
		if !model.Exist(filter) {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("该条记录不存在")))
		}
	}

	// 删除接口
	err = model.Delete(req.OperationLogIds)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("删除该改条记录失败: %s", err.Error())))
	}
	helper.Success(c, nil)
}
