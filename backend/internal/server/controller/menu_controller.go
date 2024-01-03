package controller

import (
	userLogic "micro-net-hub/internal/module/user"
	userModel "micro-net-hub/internal/module/user/model"

	"github.com/gin-gonic/gin"
)

type MenuController struct{}

// GetTree 菜单树
func (m *MenuController) GetTree(c *gin.Context) {
	req := new(userModel.MenuGetTreeReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.MenuLogicIns.GetTree(c, req)
	})
}

// GetUserMenuTreeByUserId 获取用户菜单树
func (m *MenuController) GetAccessTree(c *gin.Context) {
	req := new(userModel.MenuGetAccessTreeReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.MenuLogicIns.GetAccessTree(c, req)
	})
}

// Add 新建
func (m *MenuController) Add(c *gin.Context) {
	req := new(userModel.MenuAddReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.MenuLogicIns.Add(c, req)
	})
}

// Update 更新记录
func (m *MenuController) Update(c *gin.Context) {
	req := new(userModel.MenuUpdateReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.MenuLogicIns.Update(c, req)
	})
}

// Delete 删除记录
func (m *MenuController) Delete(c *gin.Context) {
	req := new(userModel.MenuDeleteReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.MenuLogicIns.Delete(c, req)
	})
}
