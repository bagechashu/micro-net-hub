package handler

import (
	userLogic "micro-net-hub/internal/module/user"
	userModel "micro-net-hub/internal/module/user/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct{}

// List 记录列表
func (RoleHandler) List(c *gin.Context) {
	req := new(userModel.RoleListReq)
	helper.HandleRequest(c, req, userLogic.RoleLogicIns.List)
}

// Add 新建
func (RoleHandler) Add(c *gin.Context) {
	req := new(userModel.RoleAddReq)
	helper.HandleRequest(c, req, userLogic.RoleLogicIns.Add)
}

// Update 更新记录
func (RoleHandler) Update(c *gin.Context) {
	req := new(userModel.RoleUpdateReq)
	helper.HandleRequest(c, req, userLogic.RoleLogicIns.Update)
}

// Delete 删除记录
func (RoleHandler) Delete(c *gin.Context) {
	req := new(userModel.RoleDeleteReq)
	helper.HandleRequest(c, req, userLogic.RoleLogicIns.Delete)
}

// GetMenuList 获取菜单列表
func (RoleHandler) GetMenuList(c *gin.Context) {
	req := new(userModel.RoleGetMenuListReq)
	helper.HandleRequest(c, req, userLogic.RoleLogicIns.GetMenuList)
}

// GetApiList 获取接口列表
func (RoleHandler) GetApiList(c *gin.Context) {
	req := new(userModel.RoleGetApiListReq)
	helper.HandleRequest(c, req, userLogic.RoleLogicIns.GetApiList)
}

// UpdateMenus 更新菜单
func (RoleHandler) UpdateMenus(c *gin.Context) {
	req := new(userModel.RoleUpdateMenusReq)
	helper.HandleRequest(c, req, userLogic.RoleLogicIns.UpdateMenus)
}

// UpdateApis 更新接口
func (RoleHandler) UpdateApis(c *gin.Context) {
	req := new(userModel.RoleUpdateApisReq)
	helper.HandleRequest(c, req, userLogic.RoleLogicIns.UpdateApis)
}
