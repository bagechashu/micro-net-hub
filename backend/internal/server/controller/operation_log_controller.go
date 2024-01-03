package controller

import (
	operationLogLogic "micro-net-hub/internal/module/operation_log"
	operationLogModel "micro-net-hub/internal/module/operation_log/model"

	"github.com/gin-gonic/gin"
)

type OperationLogController struct{}

// List 记录列表
func (m *OperationLogController) List(c *gin.Context) {
	req := new(operationLogModel.OperationLogListReq)
	Run(c, req, func() (interface{}, interface{}) {
		return operationLogLogic.OperationLogLogicIns.List(c, req)
	})
}

// Delete 删除记录
func (m *OperationLogController) Delete(c *gin.Context) {
	req := new(operationLogModel.OperationLogDeleteReq)
	Run(c, req, func() (interface{}, interface{}) {
		return operationLogLogic.OperationLogLogicIns.Delete(c, req)
	})
}
