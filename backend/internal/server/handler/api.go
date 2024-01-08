package handler

import (
	apimgrLogic "micro-net-hub/internal/module/apimgr"
	apiMgrModel "micro-net-hub/internal/module/apimgr/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

type ApiHandler struct{}

// List 记录列表
func (ApiHandler) List(c *gin.Context) {
	req := new(apiMgrModel.ApiListReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := apimgrLogic.List(c, req)
	helper.HandleResponse(c, data, respErr)
}

// GetTree 接口树
func (ApiHandler) GetTree(c *gin.Context) {
	req := new(apiMgrModel.ApiGetTreeReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := apimgrLogic.GetTree(c, req)
	helper.HandleResponse(c, data, respErr)
}

// Add 新建记录
func (ApiHandler) Add(c *gin.Context) {
	req := new(apiMgrModel.ApiAddReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := apimgrLogic.Add(c, req)
	helper.HandleResponse(c, data, respErr)
}

// Update 更新记录
func (ApiHandler) Update(c *gin.Context) {
	req := new(apiMgrModel.ApiUpdateReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := apimgrLogic.Update(c, req)
	helper.HandleResponse(c, data, respErr)
}

// Delete 删除记录
func (ApiHandler) Delete(c *gin.Context) {
	req := new(apiMgrModel.ApiDeleteReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := apimgrLogic.Delete(c, req)
	helper.HandleResponse(c, data, respErr)
}
