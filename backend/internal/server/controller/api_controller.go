package controller

import (
	apimgrLogic "micro-net-hub/internal/module/apimgr"
	apiMgrModel "micro-net-hub/internal/module/apimgr/model"

	"github.com/gin-gonic/gin"
)

type ApiController struct{}

// List 记录列表
func (m *ApiController) List(c *gin.Context) {
	req := new(apiMgrModel.ApiListReq)
	Run(c, req, func() (interface{}, interface{}) {
		return apimgrLogic.ApiMgrLogicIns.List(c, req)
	})
}

// GetTree 接口树
func (m *ApiController) GetTree(c *gin.Context) {
	req := new(apiMgrModel.ApiGetTreeReq)
	Run(c, req, func() (interface{}, interface{}) {
		return apimgrLogic.ApiMgrLogicIns.GetTree(c, req)
	})
}

// Add 新建记录
func (m *ApiController) Add(c *gin.Context) {
	req := new(apiMgrModel.ApiAddReq)
	Run(c, req, func() (interface{}, interface{}) {
		return apimgrLogic.ApiMgrLogicIns.Add(c, req)
	})
}

// Update 更新记录
func (m *ApiController) Update(c *gin.Context) {
	req := new(apiMgrModel.ApiUpdateReq)
	Run(c, req, func() (interface{}, interface{}) {
		return apimgrLogic.ApiMgrLogicIns.Update(c, req)
	})
}

// Delete 删除记录
func (m *ApiController) Delete(c *gin.Context) {
	req := new(apiMgrModel.ApiDeleteReq)
	Run(c, req, func() (interface{}, interface{}) {
		return apimgrLogic.ApiMgrLogicIns.Delete(c, req)
	})
}
