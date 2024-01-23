package handler

import (
	operationLogLogic "micro-net-hub/internal/module/operation_log"
	operationLogModel "micro-net-hub/internal/module/operation_log/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

type OperationLogHandler struct{}

// List 记录列表
func (OperationLogHandler) List(c *gin.Context) {
	req := new(operationLogModel.OperationLogListReq)
	helper.HandleRequest(c, req, operationLogLogic.List)
}

// Delete 删除记录
func (OperationLogHandler) Delete(c *gin.Context) {
	req := new(operationLogModel.OperationLogDeleteReq)
	helper.HandleRequest(c, req, operationLogLogic.Delete)
}
