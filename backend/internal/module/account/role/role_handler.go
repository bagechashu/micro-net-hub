package role

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/thoas/go-funk"

	"micro-net-hub/internal/global"
	"micro-net-hub/internal/module/account/current"
	"micro-net-hub/internal/module/account/model"
	apiMgrModel "micro-net-hub/internal/module/apimgr/model"
	"micro-net-hub/internal/server/helper"
	"micro-net-hub/internal/tools"
)

// List 记录列表
func RoleList(c *gin.Context) {
	req := new(model.RoleListReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*model.RoleListReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		// 获取数据列表
		var roles = model.NewRoles()
		err := roles.List(r)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取菜单列表失败: %s", err.Error()))
		}

		count, err := model.RoleCount()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取接口总数失败"))
		}

		rets := make([]model.Role, 0)
		for _, role := range roles {
			rets = append(rets, *role)
		}

		return model.RoleListRsp{
			Total: count,
			Roles: rets,
		}, nil
	})
}

// Add 新建
func RoleAdd(c *gin.Context) {
	req := new(model.RoleAddReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*model.RoleAddReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		// 获取当前用户最高角色等级
		minSort, ctxUser, err := current.GetCurrentUserMinRoleSort(c)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取当前用户最高角色等级失败: %s", err.Error()))
		}
		if minSort != 1 {
			return nil, helper.NewValidatorError(fmt.Errorf("当前用户没有权限更新角色"))
		}
		// 用户不能创建比自己等级高或相同等级的角色
		if minSort >= r.Sort {
			return nil, helper.NewValidatorError(fmt.Errorf("不能创建比自己等级高或相同等级的角色"))
		}

		role := model.Role{
			Name:    r.Name,
			Keyword: r.Keyword,
			Remark:  r.Remark,
			Status:  r.Status,
			Sort:    r.Sort,
			Creator: ctxUser.Username,
		}
		if role.Exist(tools.H{"name": r.Name}) {
			return nil, helper.NewValidatorError(fmt.Errorf("该角色名已存在"))
		}

		// 创建角色
		err = role.Add()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("创建角色失败: %s", err.Error()))
		}
		return nil, nil
	})
}

// Update 更新记录
func RoleUpdate(c *gin.Context) {
	req := new(model.RoleUpdateReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*model.RoleUpdateReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		// 当前用户角色排序最小值（最高等级角色）以及当前用户
		minSort, ctxUser, err := current.GetCurrentUserMinRoleSort(c)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取当前用户最高角色等级失败: %s", err.Error()))
		}

		if minSort != 1 {
			return nil, helper.NewValidatorError(fmt.Errorf("当前用户没有权限更新角色"))
		}

		// 不能更新比自己角色等级高或相等的角色
		// 根据path中的角色ID获取该角色信息
		roles := model.NewRoles()
		err = roles.GetRolesByIds([]uint{r.ID})
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取角色信息失败: %s", err.Error()))
		}
		if len(roles) == 0 {
			return nil, helper.NewMySqlError(fmt.Errorf("获取角色信息失败: %s", err.Error()))
		}

		if minSort >= roles[0].Sort {
			return nil, helper.NewValidatorError(fmt.Errorf("不能更新比自己角色等级高或相等的角色"))
		}

		// 不能把角色等级更新得比当前用户的等级高
		if minSort >= r.Sort {
			return nil, helper.NewValidatorError(fmt.Errorf("不能把角色等级更新得比当前用户的等级高或相同"))
		}
		filter := tools.H{"id": r.ID}
		oldData := new(model.Role)
		if !oldData.Exist(filter) {
			return nil, helper.NewValidatorError(fmt.Errorf("该角色名不存在"))
			// return nil, helper.NewMySqlError(err)
		}
		role := model.Role{
			Model:   oldData.Model,
			Name:    r.Name,
			Keyword: r.Keyword,
			Remark:  r.Remark,
			Status:  r.Status,
			Sort:    r.Sort,
			Creator: ctxUser.Username,
		}

		// 更新角色
		err = role.Update()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("更新角色失败: %s", err.Error()))
		}

		// 如果更新成功，且更新了角色的keyword, 则更新casbin中policy
		if r.Keyword != roles[0].Keyword {
			// 获取policy
			rolePolicies := global.CasbinEnforcer.GetFilteredPolicy(0, roles[0].Keyword)
			if len(rolePolicies) == 0 {
				return
			}
			rolePoliciesCopy := make([][]string, 0)
			// 替换keyword
			for _, policy := range rolePolicies {
				policyCopy := make([]string, len(policy))
				copy(policyCopy, policy)
				rolePoliciesCopy = append(rolePoliciesCopy, policyCopy)
				policy[0] = r.Keyword
			}

			//gormadapter未实现UpdatePolicies方法，等gorm更新---
			//isUpdated, _ := global.CasbinEnforcer.UpdatePolicies(rolePoliciesCopy, rolePolicies)
			//if !isUpdated {
			//	response.Fail(c, nil, "更新角色成功，但角色关键字关联的权限接口更新失败！")
			//	return
			//}

			// 这里需要先新增再删除（先删除再增加会出错）
			isAdded, _ := global.CasbinEnforcer.AddPolicies(rolePolicies)
			if !isAdded {
				return nil, helper.NewOperationError(fmt.Errorf("更新角色成功，但角色关键字关联的权限接口更新失败"))
			}
			isRemoved, _ := global.CasbinEnforcer.RemovePolicies(rolePoliciesCopy)
			if !isRemoved {
				return nil, helper.NewOperationError(fmt.Errorf("更新角色成功，但角色关键字关联的权限接口更新失败"))
			}
			err := global.CasbinEnforcer.LoadPolicy()
			if err != nil {
				return nil, helper.NewOperationError(fmt.Errorf("更新角色成功，但角色关键字关联角色的权限接口策略加载失败"))
			}

		}

		// 更新角色成功处理用户信息缓存有两种做法:（这里使用第二种方法，因为一个角色下用户数量可能很多，第二种方法可以分散数据库压力）
		// 1.可以帮助用户更新拥有该角色的用户信息缓存,使用下面方法
		// err = ur.UpdateUserInfoCacheByRoleId(uint(roleId))
		// 2.直接清理缓存，让活跃的用户自己重新缓存最新用户信息
		model.ClearUserInfoCache()

		return nil, nil
	})
}

// Delete 删除记录
func RoleDelete(c *gin.Context) {
	req := new(model.RoleDeleteReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*model.RoleDeleteReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		// 获取当前登陆用户最高等级角色
		minSort, _, err := current.GetCurrentUserMinRoleSort(c)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取当前用户最高角色等级失败: %s", err.Error()))
		}

		// 获取角色信息
		roles := model.NewRoles()
		err = roles.GetRolesByIds(r.RoleIds)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取角色信息失败: %s", err.Error()))
		}
		if len(roles) == 0 {
			return nil, helper.NewMySqlError(fmt.Errorf("未能获取到角色信息"))
		}

		// 不能删除比自己角色等级高或相等的角色
		for _, role := range roles {
			if minSort >= role.Sort {
				return nil, helper.NewValidatorError(fmt.Errorf("不能删除比自己角色等级高或相等的角色"))
			}
		}

		// 删除角色
		err = roles.Delete()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("删除角色失败: %s", err.Error()))
		}

		// 删除角色成功直接清理缓存，让活跃的用户自己重新缓存最新用户信息
		model.ClearUserInfoCache()
		return nil, nil
	})
}

// RoleGetMenuList 获取角色菜单列表
func RoleGetMenuList(c *gin.Context) {
	req := new(model.RoleGetMenuListReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*model.RoleGetMenuListReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		menus, err := model.GetRoleMenusById(r.RoleID)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取角色的权限菜单失败: " + err.Error()))
		}
		return menus, nil
	})
}

// RoleGetApiList 获取角色接口列表
func RoleGetApiList(c *gin.Context) {
	req := new(model.RoleGetApiListReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*model.RoleGetApiListReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		role := new(model.Role)
		err := role.Find(tools.H{"id": r.RoleID})
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取资源失败: " + err.Error()))
		}

		policies := global.CasbinEnforcer.GetFilteredPolicy(0, role.Keyword)

		apis, err := apiMgrModel.ListAll()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取资源列表失败: " + err.Error()))
		}
		accessApis := make([]*apiMgrModel.Api, 0)

		for _, policy := range policies {
			path := policy[1]
			method := policy[2]
			for _, api := range apis {
				if path == api.Path && method == api.Method {
					accessApis = append(accessApis, api)
					break
				}
			}
		}

		return accessApis, nil
	})
}

// RoleUpdateMenus 更新角色菜单
func RoleUpdateMenus(c *gin.Context) {
	req := new(model.RoleUpdateMenusReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*model.RoleUpdateMenusReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		roles := model.NewRoles()
		err := roles.GetRolesByIds([]uint{r.RoleID})
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取角色信息失败: %s", err.Error()))
		}
		if len(roles) == 0 {
			return nil, helper.NewMySqlError(fmt.Errorf("未获取到角色信息"))
		}

		// 当前用户角色排序最小值（最高等级角色）以及当前用户
		minSort, ctxUser, err := current.GetCurrentUserMinRoleSort(c)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取当前用户最高角色等级失败: %s", err.Error()))
		}

		// TODO: roles[0] ?????
		global.Log.Infof("所有角色列表: %+v", roles)
		// (非管理员)不能更新比自己角色等级高或相等角色的权限菜单
		if minSort != 1 {
			if minSort >= roles[0].Sort {
				return nil, helper.NewValidatorError(fmt.Errorf("不能更新比自己角色等级高或相等角色的权限菜单"))
			}
		}

		// 获取当前用户所拥有的权限菜单
		ctxUserMenus, err := model.GetUserMenusByUserId(ctxUser.ID)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取当前用户的可访问菜单列表失败: " + err.Error()))
		}

		// 获取当前用户所拥有的权限菜单ID
		ctxUserMenusIds := make([]uint, 0)
		for _, menu := range ctxUserMenus {
			ctxUserMenusIds = append(ctxUserMenusIds, menu.ID)
		}

		// 用户需要修改的菜单集合
		reqMenus := make([]*model.Menu, 0)

		// (非管理员)不能把角色的权限菜单设置的比当前用户所拥有的权限菜单多
		if minSort != 1 {
			for _, id := range r.MenuIds {
				if !funk.Contains(ctxUserMenusIds, id) {
					return nil, helper.NewValidatorError(fmt.Errorf("无权设置ID为%d的菜单", id))
				}
			}

			for _, id := range r.MenuIds {
				for _, menu := range ctxUserMenus {
					if id == menu.ID {
						reqMenus = append(reqMenus, menu)
						break
					}
				}
			}
		} else {
			// 管理员随意设置
			// 根据menuIds查询查询菜单
			var menus = model.NewMenus()
			err := menus.List()
			if err != nil {
				return nil, helper.NewValidatorError(fmt.Errorf("获取菜单列表失败: " + err.Error()))
			}
			for _, menuId := range r.MenuIds {
				for _, menu := range menus {
					if menuId == menu.ID {
						reqMenus = append(reqMenus, menu)
					}
				}
			}
		}

		roles[0].Menus = reqMenus

		err = roles[0].UpdateRoleMenus()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("更新角色的权限菜单失败: " + err.Error()))
		}

		return nil, nil
	})
}

// RoleUpdateApis 更新角色接口
func RoleUpdateApis(c *gin.Context) {
	req := new(model.RoleUpdateApisReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*model.RoleUpdateApisReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		// 根据path中的角色ID获取该角色信息
		roles := model.NewRoles()
		if err := roles.GetRolesByIds([]uint{r.RoleID}); err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取角色信息失败: " + err.Error()))
		}
		if len(roles) == 0 {
			return nil, helper.NewMySqlError(fmt.Errorf("未获取到角色信息"))
		}

		// 当前用户角色排序最小值（最高等级角色）以及当前用户
		minSort, ctxUser, err := current.GetCurrentUserMinRoleSort(c)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取当前用户最高角色等级失败: %s", err.Error()))
		}

		// (非管理员)不能更新比自己角色等级高或相等角色的权限菜单
		if minSort != 1 {
			if minSort >= roles[0].Sort {
				return nil, helper.NewValidatorError(fmt.Errorf("不能更新比自己角色等级高或相等角色的权限菜单"))
			}
		}

		// 获取当前用户所拥有的权限接口
		ctxRoles := ctxUser.Roles
		ctxRolesPolicies := make([][]string, 0)
		for _, role := range ctxRoles {
			policy := global.CasbinEnforcer.GetFilteredPolicy(0, role.Keyword)
			ctxRolesPolicies = append(ctxRolesPolicies, policy...)
		}
		// 得到path中的角色ID对应角色能够设置的权限接口集合
		for _, policy := range ctxRolesPolicies {
			policy[0] = roles[0].Keyword
		}

		// 根据apiID获取接口详情
		apis, err := apiMgrModel.GetApisById(r.ApiIds)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("根据接口ID获取接口信息失败"))
		}
		// 生成前端想要设置的角色policies
		reqRolePolicies := make([][]string, 0)
		for _, api := range apis {
			reqRolePolicies = append(reqRolePolicies, []string{
				roles[0].Keyword, api.Path, api.Method,
			})
		}

		// (非管理员)不能把角色的权限接口设置的比当前用户所拥有的权限接口多
		if minSort != 1 {
			for _, reqPolicy := range reqRolePolicies {
				if !funk.Contains(ctxRolesPolicies, reqPolicy) {
					return nil, helper.NewValidatorError(fmt.Errorf("无权设置路径为%s,请求方式为%s的接口", reqPolicy[1], reqPolicy[2]))
				}
			}
		}

		// 更新角色的权限接口
		err = model.UpdateRoleApis(roles[0].Keyword, reqRolePolicies)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("更新角色的权限接口失败"))
		}
		return nil, nil
	})
}
