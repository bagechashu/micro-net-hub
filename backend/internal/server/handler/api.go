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
	helper.HandleRequest(c, req, apimgrLogic.List)
}

// GetTree 接口树
func (ApiHandler) GetTree(c *gin.Context) {
	req := new(apiMgrModel.ApiGetTreeReq)
	helper.HandleRequest(c, req, apimgrLogic.GetTree)
}

// Add 新建记录
func (ApiHandler) Add(c *gin.Context) {
	req := new(apiMgrModel.ApiAddReq)
	helper.HandleRequest(c, req, apimgrLogic.Add)
}

// Update 更新记录
func (ApiHandler) Update(c *gin.Context) {
	req := new(apiMgrModel.ApiUpdateReq)
	helper.HandleRequest(c, req, apimgrLogic.Update)
}

// Delete 删除记录
func (ApiHandler) Delete(c *gin.Context) {
	req := new(apiMgrModel.ApiDeleteReq)
	helper.HandleRequest(c, req, apimgrLogic.Delete)
}
