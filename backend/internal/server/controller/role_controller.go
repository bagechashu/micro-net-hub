package controller

import (
	userLogic "micro-net-hub/internal/module/user"
	userModel "micro-net-hub/internal/module/user/model"

	"github.com/gin-gonic/gin"
)

type RoleController struct{}

// List 记录列表
func (m *RoleController) List(c *gin.Context) {
	req := new(userModel.RoleListReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.RoleLogicIns.List(c, req)
	})
}

// Add 新建
func (m *RoleController) Add(c *gin.Context) {
	req := new(userModel.RoleAddReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.RoleLogicIns.Add(c, req)
	})
}

// Update 更新记录
func (m *RoleController) Update(c *gin.Context) {
	req := new(userModel.RoleUpdateReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.RoleLogicIns.Update(c, req)
	})
}

// Delete 删除记录
func (m *RoleController) Delete(c *gin.Context) {
	req := new(userModel.RoleDeleteReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.RoleLogicIns.Delete(c, req)
	})
}

// GetMenuList 获取菜单列表
func (m *RoleController) GetMenuList(c *gin.Context) {
	req := new(userModel.RoleGetMenuListReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.RoleLogicIns.GetMenuList(c, req)
	})
}

// GetApiList 获取接口列表
func (m *RoleController) GetApiList(c *gin.Context) {
	req := new(userModel.RoleGetApiListReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.RoleLogicIns.GetApiList(c, req)
	})
}

// UpdateMenus 更新菜单
func (m *RoleController) UpdateMenus(c *gin.Context) {
	req := new(userModel.RoleUpdateMenusReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.RoleLogicIns.UpdateMenus(c, req)
	})
}

// UpdateApis 更新接口
func (m *RoleController) UpdateApis(c *gin.Context) {
	req := new(userModel.RoleUpdateApisReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.RoleLogicIns.UpdateApis(c, req)
	})
}
