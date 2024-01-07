package user

import (
	"fmt"

	"micro-net-hub/internal/server/helper"
	"micro-net-hub/internal/tools"

	userModel "micro-net-hub/internal/module/user/model"

	"github.com/gin-gonic/gin"
)

type MenuLogic struct{}

// Add 添加数据
func (l MenuLogic) Add(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.MenuAddReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c

	// 获取当前用户
	ctxUser, err := GetCurrentLoginUser(c)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败"))
	}

	var menu = &userModel.Menu{
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
}

// // List 数据列表
// func (l MenuLogic) List(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
// 	_, ok := req.(*request.MenuListReq)
// 	if !ok {
// 		return nil, helper.ReqAssertErr
// 	}
// 	_ = c

// 	menus, err := userModel.MenuSrvIns.List()
// 	if err != nil {
// 		return nil, helper.NewMySqlError(fmt.Errorf("获取资源列表失败: %s", err.Error()))
// 	}

// 	rets := make([]model.Menu, 0)
// 	for _, menu := range menus {
// 		rets = append(rets, *menu)
// 	}
// 	count, err := userModel.MenuSrvIns.Count()
// 	if err != nil {
// 		return nil, helper.NewMySqlError(fmt.Errorf("获取资源总数失败"))
// 	}

// 	return response.MenuListRsp{
// 		Total: count,
// 		Menus: rets,
// 	}, nil
// }

// Update 更新数据
func (l MenuLogic) Update(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.MenuUpdateReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c

	oldMenu := new(userModel.Menu)
	filter := tools.H{"id": int(r.ID)}
	if !oldMenu.Exist(filter) {
		return nil, helper.NewMySqlError(fmt.Errorf("该ID对应的记录不存在"))
	}

	// 获取当前登陆用户
	ctxUser, err := GetCurrentLoginUser(c)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户失败"))
	}

	newMenu := &userModel.Menu{
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
}

// Delete 删除数据
func (l MenuLogic) Delete(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.MenuDeleteReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c

	for _, id := range r.MenuIds {
		filter := tools.H{"id": int(id)}
		var menu = new(userModel.Menu)
		if !menu.Exist(filter) {
			return nil, helper.NewMySqlError(fmt.Errorf("Menu ID: %d 对应的记录不存在", menu.ID))
		}
		// 删除接口
		err := menu.Delete()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("%s删除接口失败: %s", menu.Name, err.Error()))
		}
	}
	return nil, nil
}

// GetTree 数据树
func (l MenuLogic) GetTree(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	_, ok := req.(*userModel.MenuGetTreeReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c
	var menus = userModel.NewMenus()
	err := menus.List()
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取资源列表失败: " + err.Error()))
	}

	tree := userModel.GenMenuTree(0, menus)

	return tree, nil
}

// GetAccessTree 获取用户菜单树
func (l MenuLogic) GetAccessTree(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.MenuGetAccessTreeReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c
	// 校验
	filter := tools.H{"id": r.ID}
	var u userModel.User
	if !u.Exist(filter) {
		return nil, helper.NewValidatorError(fmt.Errorf("该用户不存在"))
	}
	user := new(userModel.User)
	err := user.Find(filter)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("在MySQL查询用户失败: " + err.Error()))
	}
	var roleIds []uint
	for _, role := range user.Roles {
		roleIds = append(roleIds, role.ID)
	}
	var menus = userModel.NewMenus()
	err = menus.ListUserMenus(roleIds)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取资源列表失败: " + err.Error()))
	}

	tree := userModel.GenMenuTree(0, menus)

	return tree, nil
}
