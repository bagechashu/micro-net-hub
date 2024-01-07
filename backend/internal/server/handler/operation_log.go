package handler

import (
	operationLogLogic "micro-net-hub/internal/module/operation_log"
	operationLogModel "micro-net-hub/internal/module/operation_log/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

type OperationLogController struct{}

// List 记录列表
func (m *OperationLogController) List(c *gin.Context) {
	req := new(operationLogModel.OperationLogListReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := operationLogLogic.List(c, req)
	helper.HandleResponse(c, data, respErr)
}

// Delete 删除记录
func (m *OperationLogController) Delete(c *gin.Context) {
	req := new(operationLogModel.OperationLogDeleteReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := operationLogLogic.Delete(c, req)
	helper.HandleResponse(c, data, respErr)
}
