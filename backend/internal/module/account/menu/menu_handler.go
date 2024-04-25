package menu

import (
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"

	"fmt"

	"micro-net-hub/internal/module/account/current"
	accountModel "micro-net-hub/internal/module/account/model"
	"micro-net-hub/internal/tools"
)

// GetTree 菜单树
func GetTree(c *gin.Context) {
	req := new(accountModel.MenuGetTreeReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		_, ok := req.(*accountModel.MenuGetTreeReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c
		var menus = accountModel.NewMenus()
		err := menus.List()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取资源列表失败: " + err.Error()))
		}

		tree := accountModel.GenMenuTree(0, menus)

		return tree, nil
	})
}

// GetUserMenuTreeByUserId 获取用户菜单树
func GetAccessTree(c *gin.Context) {
	req := new(accountModel.MenuGetAccessTreeReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.MenuGetAccessTreeReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c
		// 校验
		filter := tools.H{"id": r.ID}
		var u accountModel.User
		if !u.Exist(filter) {
			return nil, helper.NewValidatorError(fmt.Errorf("该用户不存在"))
		}
		user := new(accountModel.User)
		err := user.Find(filter)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("在MySQL查询用户失败: " + err.Error()))
		}
		var roleIds []uint
		for _, role := range user.Roles {
			roleIds = append(roleIds, role.ID)
		}
		var menus = accountModel.NewMenus()
		err = menus.ListUserMenus(roleIds)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取资源列表失败: " + err.Error()))
		}

		tree := accountModel.GenMenuTree(0, menus)

		return tree, nil
	})
}

// Add 新建
func Add(c *gin.Context) {
	req := new(accountModel.MenuAddReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.MenuAddReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		// 获取当前用户
		ctxUser, err := current.GetCurrentLoginUser(c)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败"))
		}

		var menu = &accountModel.Menu{
			Name:       r.Name,
			Title:      r.Title,
			Icon:       r.Icon,
			Path:       r.Path,
			Redirect:   r.Redirect,
			Component:  r.Component,
			Sort:       r.Sort,
			Status:     r.Status,
			Hidden:     r.Hidden,
			NoCache:    r.NoCache,
			AlwaysShow: r.AlwaysShow,
			Breadcrumb: r.Breadcrumb,
			ActiveMenu: r.ActiveMenu,
			ParentId:   r.ParentId,
			Creator:    ctxUser.Username,
		}
		if menu.Exist(tools.H{"name": r.Name}) {
			return nil, helper.NewMySqlError(fmt.Errorf("菜单名称已存在"))

		}
		err = menu.Add()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("创建记录失败: %s", err.Error()))
		}

		return nil, nil
	})
}

// Update 更新记录
func Update(c *gin.Context) {
	req := new(accountModel.MenuUpdateReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.MenuUpdateReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		oldMenu := new(accountModel.Menu)
		filter := tools.H{"id": int(r.ID)}
		if !oldMenu.Exist(filter) {
			return nil, helper.NewMySqlError(fmt.Errorf("该ID对应的记录不存在"))
		}

		// 获取当前登陆用户
		ctxUser, err := current.GetCurrentLoginUser(c)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户失败"))
		}

		newMenu := &accountModel.Menu{
			Model:      oldMenu.Model,
			Name:       r.Name,
			Title:      r.Title,
			Icon:       r.Icon,
			Path:       r.Path,
			Redirect:   r.Redirect,
			Component:  r.Component,
			Sort:       r.Sort,
			Status:     r.Status,
			Hidden:     r.Hidden,
			NoCache:    r.NoCache,
			AlwaysShow: r.AlwaysShow,
			Breadcrumb: r.Breadcrumb,
			ActiveMenu: r.ActiveMenu,
			ParentId:   r.ParentId,
			Creator:    ctxUser.Username,
		}
		err = newMenu.Update()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("更新记录失败: %s", err.Error()))
		}

		return nil, nil
	})
}

// Delete 删除记录
func Delete(c *gin.Context) {
	req := new(accountModel.MenuDeleteReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.MenuDeleteReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		for _, id := range r.MenuIds {
			filter := tools.H{"id": int(id)}
			var menu = new(accountModel.Menu)
			if !menu.Exist(filter) {
				return nil, helper.NewMySqlError(fmt.Errorf("menu ID: %d 对应的记录不存在", menu.ID))
			}
			// 删除接口
			err := menu.Delete()
			if err != nil {
				return nil, helper.NewMySqlError(fmt.Errorf("%s删除接口失败: %s", menu.Name, err.Error()))
			}
		}
		return nil, nil
	})
}
