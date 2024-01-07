package handler

import (
	userLogic "micro-net-hub/internal/module/user"
	userModel "micro-net-hub/internal/module/user/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

type MenuController struct{}

// GetTree 菜单树
func (m *MenuController) GetTree(c *gin.Context) {
	req := new(userModel.MenuGetTreeReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.MenuLogicIns.GetTree(c, req)
	helper.HandleResponse(c, data, respErr)
}

// GetUserMenuTreeByUserId 获取用户菜单树
func (m *MenuController) GetAccessTree(c *gin.Context) {
	req := new(userModel.MenuGetAccessTreeReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.MenuLogicIns.GetAccessTree(c, req)
	helper.HandleResponse(c, data, respErr)
}

// Add 新建
func (m *MenuController) Add(c *gin.Context) {
	req := new(userModel.MenuAddReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.MenuLogicIns.Add(c, req)
	helper.HandleResponse(c, data, respErr)
}

// Update 更新记录
func (m *MenuController) Update(c *gin.Context) {
	req := new(userModel.MenuUpdateReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.MenuLogicIns.Update(c, req)
	helper.HandleResponse(c, data, respErr)
}

// Delete 删除记录
func (m *MenuController) Delete(c *gin.Context) {
	req := new(userModel.MenuDeleteReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.MenuLogicIns.Delete(c, req)
	helper.HandleResponse(c, data, respErr)
}
