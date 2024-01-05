package controller

import (
	fieldRelationLogic "micro-net-hub/internal/module/goldap/field_relation"
	fieldRelationModel "micro-net-hub/internal/module/goldap/field_relation/model"

	"github.com/gin-gonic/gin"
)

type FieldRelationController struct{}

// List 记录列表
func (m *FieldRelationController) List(c *gin.Context) {
	req := new(fieldRelationModel.FieldRelationListReq)
	Run(c, req, func() (interface{}, interface{}) {
		return fieldRelationLogic.List(c, req)
	})
}

// Add 新建记录
func (m *FieldRelationController) Add(c *gin.Context) {
	req := new(fieldRelationModel.FieldRelationAddReq)
	Run(c, req, func() (interface{}, interface{}) {
		return fieldRelationLogic.Add(c, req)
	})
}

// Update 更新记录
func (m *FieldRelationController) Update(c *gin.Context) {
	req := new(fieldRelationModel.FieldRelationUpdateReq)
	Run(c, req, func() (interface{}, interface{}) {
		return fieldRelationLogic.Update(c, req)
	})
}

// Delete 删除记录
func (m *FieldRelationController) Delete(c *gin.Context) {
	req := new(fieldRelationModel.FieldRelationDeleteReq)
	Run(c, req, func() (interface{}, interface{}) {
		return fieldRelationLogic.Delete(c, req)
	})
}
