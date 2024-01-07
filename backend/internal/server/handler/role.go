package handler

import (
	userLogic "micro-net-hub/internal/module/user"
	userModel "micro-net-hub/internal/module/user/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

type RoleController struct{}

// List 记录列表
func (m *RoleController) List(c *gin.Context) {
	req := new(userModel.RoleListReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.RoleLogicIns.List(c, req)
	helper.HandleResponse(c, data, respErr)
}

// Add 新建
func (m *RoleController) Add(c *gin.Context) {
	req := new(userModel.RoleAddReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.RoleLogicIns.Add(c, req)
	helper.HandleResponse(c, data, respErr)
}

// Update 更新记录
func (m *RoleController) Update(c *gin.Context) {
	req := new(userModel.RoleUpdateReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.RoleLogicIns.Update(c, req)
	helper.HandleResponse(c, data, respErr)
}

// Delete 删除记录
func (m *RoleController) Delete(c *gin.Context) {
	req := new(userModel.RoleDeleteReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.RoleLogicIns.Delete(c, req)
	helper.HandleResponse(c, data, respErr)
}

// GetMenuList 获取菜单列表
func (m *RoleController) GetMenuList(c *gin.Context) {
	req := new(userModel.RoleGetMenuListReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.RoleLogicIns.GetMenuList(c, req)
	helper.HandleResponse(c, data, respErr)
}

// GetApiList 获取接口列表
func (m *RoleController) GetApiList(c *gin.Context) {
	req := new(userModel.RoleGetApiListReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.RoleLogicIns.GetApiList(c, req)
	helper.HandleResponse(c, data, respErr)
}

// UpdateMenus 更新菜单
func (m *RoleController) UpdateMenus(c *gin.Context) {
	req := new(userModel.RoleUpdateMenusReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.RoleLogicIns.UpdateMenus(c, req)
	helper.HandleResponse(c, data, respErr)
}

// UpdateApis 更新接口
func (m *RoleController) UpdateApis(c *gin.Context) {
	req := new(userModel.RoleUpdateApisReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.RoleLogicIns.UpdateApis(c, req)
	helper.HandleResponse(c, data, respErr)
}
