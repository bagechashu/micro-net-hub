package handler

import (
	fieldRelationLogic "micro-net-hub/internal/module/goldap/field_relation"
	fieldRelationModel "micro-net-hub/internal/module/goldap/field_relation/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

type FieldRelationHandler struct{}

// List 记录列表
func (FieldRelationHandler) List(c *gin.Context) {
	req := new(fieldRelationModel.FieldRelationListReq)
	helper.HandleRequest(c, req, fieldRelationLogic.List)
}

// Add 新建记录
func (FieldRelationHandler) Add(c *gin.Context) {
	req := new(fieldRelationModel.FieldRelationAddReq)
	helper.HandleRequest(c, req, fieldRelationLogic.Add)
}

// Update 更新记录
func (FieldRelationHandler) Update(c *gin.Context) {
	req := new(fieldRelationModel.FieldRelationUpdateReq)
	helper.HandleRequest(c, req, fieldRelationLogic.Update)
}

// Delete 删除记录
func (FieldRelationHandler) Delete(c *gin.Context) {
	req := new(fieldRelationModel.FieldRelationDeleteReq)
	helper.HandleRequest(c, req, fieldRelationLogic.Delete)
}
