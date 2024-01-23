package handler

import (
	userLogic "micro-net-hub/internal/module/user"
	userModel "micro-net-hub/internal/module/user/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

type MenuHandler struct{}

// GetTree 菜单树
func (MenuHandler) GetTree(c *gin.Context) {
	req := new(userModel.MenuGetTreeReq)
	helper.HandleRequest(c, req, userLogic.MenuLogicIns.GetTree)
}

// GetUserMenuTreeByUserId 获取用户菜单树
func (MenuHandler) GetAccessTree(c *gin.Context) {
	req := new(userModel.MenuGetAccessTreeReq)
	helper.HandleRequest(c, req, userLogic.MenuLogicIns.GetAccessTree)
}

// Add 新建
func (MenuHandler) Add(c *gin.Context) {
	req := new(userModel.MenuAddReq)
	helper.HandleRequest(c, req, userLogic.MenuLogicIns.Add)
}

// Update 更新记录
func (MenuHandler) Update(c *gin.Context) {
	req := new(userModel.MenuUpdateReq)
	helper.HandleRequest(c, req, userLogic.MenuLogicIns.Update)
}

// Delete 删除记录
func (MenuHandler) Delete(c *gin.Context) {
	req := new(userModel.MenuDeleteReq)
	helper.HandleRequest(c, req, userLogic.MenuLogicIns.Delete)
}
