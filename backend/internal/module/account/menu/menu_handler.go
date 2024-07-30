package menu

import (
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"

	"fmt"

	"micro-net-hub/internal/module/account/auth"
	accountModel "micro-net-hub/internal/module/account/model"
)

// GetTree 菜单树
func GetTree(c *gin.Context) {
	var menus = accountModel.NewMenus()
	err := menus.List()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取资源列表失败: %s", err.Error())))
		return
	}

	tree := accountModel.GenMenuTree(0, menus)
	helper.Success(c, tree)
}

// MenuGetAccessTreeReq 用户菜单树request struct
type MenuGetAccessTreeReq struct {
	ID uint `json:"id" form:"id"`
}

// GetUserMenuTreeByUserId 获取用户菜单树
func GetAccessTree(c *gin.Context) {
	var req MenuGetAccessTreeReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 校验
	filter := map[string]interface{}{"id": req.ID}
	var u accountModel.User
	if !u.Exist(filter) {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("该用户不存在")))
		return
	}
	user := new(accountModel.User)
	err = user.Find(filter)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("查询用户失败: %s", err.Error())))
		return
	}
	var roleIds []uint
	for _, role := range user.Roles {
		roleIds = append(roleIds, role.ID)
	}
	var menus = accountModel.NewMenus()
	err = menus.ListUserMenus(roleIds)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取资源列表失败: %s", err.Error())))
		return
	}

	tree := accountModel.GenMenuTree(0, menus)

	helper.Success(c, tree)
}

// MenuAddReq 添加资源结构体
type MenuAddReq struct {
	Name       string                                `json:"name" validate:"required,min=1,max=50"`
	Icon       string                                `json:"icon" validate:"min=0,max=50"`
	Path       string                                `json:"path" validate:"required,min=1,max=100"`
	Redirect   string                                `json:"redirect" validate:"min=0,max=100"`
	Component  string                                `json:"component" validate:"required,min=1,max=100"`
	Sort       uint                                  `json:"sort" validate:"gte=1,lte=999"`
	Status     accountModel.MenuStatus               `json:"status" validate:"oneof=1 2"`
	Hidden     accountModel.MenuVisibility           `json:"hidden" validate:"oneof=1 2"`
	NoCache    accountModel.MenuCachePolicy          `json:"noCache" validate:"oneof=1 2"`
	AlwaysShow accountModel.MenuRootRouteDisplay     `json:"alwaysShow" validate:"oneof=1 2"`
	Breadcrumb accountModel.MenuBreadcrumbVisibility `json:"breadcrumb" validate:"oneof=1 2"`
	ActiveMenu string                                `json:"activeMenu" validate:"min=0,max=100"`
	ParentId   uint                                  `json:"parentId"`
}

// Add 新建
func Add(c *gin.Context) {
	var req MenuAddReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取当前用户
	ctxUser, err := auth.GetCtxLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败: %s", err.Error())))
		return
	}

	var menu = &accountModel.Menu{
		Name:       req.Name,
		Icon:       req.Icon,
		Path:       req.Path,
		Redirect:   req.Redirect,
		Component:  req.Component,
		Sort:       req.Sort,
		Status:     req.Status,
		Hidden:     req.Hidden,
		NoCache:    req.NoCache,
		AlwaysShow: req.AlwaysShow,
		Breadcrumb: req.Breadcrumb,
		ActiveMenu: req.ActiveMenu,
		ParentId:   req.ParentId,
		Creator:    ctxUser.Username,
	}
	if menu.Exist(map[string]interface{}{"name": req.Name}) {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("菜单名称已存在: %s", req.Name)))
		return
	}
	err = menu.Add()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("创建记录失败: %s", err.Error())))
		return
	}

	helper.Success(c, nil)
}

// MenuUpdateReq 更新资源结构体
type MenuUpdateReq struct {
	ID         uint                                  `json:"id" validate:"required"`
	Name       string                                `json:"name" validate:"required,min=1,max=50"`
	Icon       string                                `json:"icon" validate:"min=0,max=50"`
	Path       string                                `json:"path" validate:"required,min=1,max=100"`
	Redirect   string                                `json:"redirect" validate:"min=0,max=100"`
	Component  string                                `json:"component" validate:"min=0,max=100"`
	Sort       uint                                  `json:"sort" validate:"gte=1,lte=999"`
	Status     accountModel.MenuStatus               `json:"status" validate:"oneof=1 2"`
	Hidden     accountModel.MenuVisibility           `json:"hidden" validate:"oneof=1 2"`
	NoCache    accountModel.MenuCachePolicy          `json:"noCache" validate:"oneof=1 2"`
	AlwaysShow accountModel.MenuRootRouteDisplay     `json:"alwaysShow" validate:"oneof=1 2"`
	Breadcrumb accountModel.MenuBreadcrumbVisibility `json:"breadcrumb" validate:"oneof=1 2"`
	ActiveMenu string                                `json:"activeMenu" validate:"min=0,max=100"`
	ParentId   uint                                  `json:"parentId" validate:"min=0,max=1000"`
}

// Update 更新记录
func Update(c *gin.Context) {
	var req MenuUpdateReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	oldMenu := new(accountModel.Menu)
	filter := map[string]interface{}{"id": int(req.ID)}
	if !oldMenu.Exist(filter) {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("该ID对应的记录不存在: %d", req.ID)))
		return
	}

	// 获取当前登陆用户
	ctxUser, err := auth.GetCtxLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户失败: %s", err.Error())))
	}

	newMenu := &accountModel.Menu{
		Model:      oldMenu.Model,
		Name:       req.Name,
		Icon:       req.Icon,
		Path:       req.Path,
		Redirect:   req.Redirect,
		Component:  req.Component,
		Sort:       req.Sort,
		Status:     req.Status,
		Hidden:     req.Hidden,
		NoCache:    req.NoCache,
		AlwaysShow: req.AlwaysShow,
		Breadcrumb: req.Breadcrumb,
		ActiveMenu: req.ActiveMenu,
		ParentId:   req.ParentId,
		Creator:    ctxUser.Username,
	}
	err = newMenu.Update()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("更新记录失败: %s", err.Error())))
		return
	}

	helper.Success(c, nil)
}

// MenuDeleteReq 删除资源结构体
type MenuDeleteReq struct {
	MenuIds []uint `json:"menuIds" validate:"required"`
}

// Delete 删除记录
func Delete(c *gin.Context) {
	var req MenuDeleteReq

	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	for _, id := range req.MenuIds {
		filter := map[string]interface{}{"id": int(id)}
		var menu = new(accountModel.Menu)
		if !menu.Exist(filter) {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("menu ID: %d 对应的记录不存在", menu.ID)))
			return
		}
		// 删除接口
		err := menu.Delete()
		if err != nil {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("%s删除接口失败: %s", menu.Name, err.Error())))
			return
		}
	}

	helper.Success(c, nil)
}
