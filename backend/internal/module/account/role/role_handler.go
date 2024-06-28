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
)

// RoleListReq 列表结构体
type RoleListReq struct {
	Name     string `json:"name" form:"name"`
	Keyword  string `json:"keyword" form:"keyword"`
	Status   uint   `json:"status" form:"status"`
	PageNum  int    `json:"pageNum" form:"pageNum"`
	PageSize int    `json:"pageSize" form:"pageSize"`
}

type RoleListRsp struct {
	Total int64        `json:"total"`
	Roles []model.Role `json:"roles"`
}

// List 记录列表
func RoleList(c *gin.Context) {
	var req RoleListReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取数据列表
	var roles = model.NewRoles()
	err = roles.List(
		&model.Role{
			Name:    req.Name,
			Keyword: req.Keyword,
			Status:  req.Status,
		},
		req.PageNum,
		req.PageSize,
	)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取菜单列表失败: %s", err.Error())))
		return
	}

	count, err := model.RoleCount()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取接口总数失败")))
		return
	}

	rets := make([]model.Role, 0)
	for _, role := range roles {
		rets = append(rets, *role)
	}

	helper.Success(c, RoleListRsp{
		Total: count,
		Roles: rets,
	})
}

// RoleAddReq 添加资源结构体
type RoleAddReq struct {
	Name    string `json:"name" validate:"required,min=1,max=20"`
	Keyword string `json:"keyword" validate:"required,min=1,max=20"`
	Remark  string `json:"remark" validate:"min=0,max=100"`
	Status  uint   `json:"status" validate:"oneof=1 2"`
	Sort    uint   `json:"sort" validate:"gte=1,lte=999"`
}

// Add 新建
func RoleAdd(c *gin.Context) {
	var req RoleAddReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取当前用户最高角色等级
	minSort, ctxUser, err := current.GetCurrentUserMinRoleSort(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前用户最高角色等级失败: %s", err.Error())))
		return
	}
	if minSort != 1 {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("当前用户没有权限更新角色")))
		return
	}
	// 用户不能创建比自己等级高或相同等级的角色
	if minSort >= req.Sort {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("不能创建比自己等级高或相同等级的角色")))
		return
	}

	role := model.Role{
		Name:    req.Name,
		Keyword: req.Keyword,
		Remark:  req.Remark,
		Status:  req.Status,
		Sort:    req.Sort,
		Creator: ctxUser.Username,
	}
	if role.Exist(map[string]interface{}{"name": req.Name}) {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("该角色名已存在")))
		return
	}

	// 创建角色
	err = role.Add()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("创建角色失败: %s", err.Error())))
		return
	}
	helper.Success(c, nil)
}

// RoleUpdateReq 更新资源结构体
type RoleUpdateReq struct {
	ID      uint   `json:"id" validate:"required"`
	Name    string `json:"name" validate:"required,min=1,max=20"`
	Keyword string `json:"keyword" validate:"required,min=1,max=20"`
	Remark  string `json:"remark" validate:"min=0,max=100"`
	Status  uint   `json:"status" validate:"oneof=1 2"`
	Sort    uint   `json:"sort" validate:"gte=1,lte=999"`
}

// Update 更新记录
func RoleUpdate(c *gin.Context) {
	var req RoleUpdateReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 当前用户角色排序最小值（最高等级角色）以及当前用户
	minSort, ctxUser, err := current.GetCurrentUserMinRoleSort(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前用户最高角色等级失败: %s", err.Error())))
		return
	}

	if minSort != 1 {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("当前用户没有权限更新角色")))
		return
	}

	// 不能更新比自己角色等级高或相等的角色
	// 根据path中的角色ID获取该角色信息
	roles := model.NewRoles()
	err = roles.GetRolesByIds([]uint{req.ID})
	if err != nil || len(roles) == 0 {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取角色信息失败: %s", err.Error())))
		return
	}

	if minSort >= roles[0].Sort {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("不能更新比自己角色等级高或相等的角色")))
		return
	}

	// 不能把角色等级更新得比当前用户的等级高
	if minSort >= req.Sort {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("不能把角色等级更新得比当前用户的等级高或相同")))
		return
	}
	filter := map[string]interface{}{"id": req.ID}
	oldData := new(model.Role)
	if !oldData.Exist(filter) {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("该角色名不存在")))
		return
	}
	role := model.Role{
		Model:   oldData.Model,
		Name:    req.Name,
		Keyword: req.Keyword,
		Remark:  req.Remark,
		Status:  req.Status,
		Sort:    req.Sort,
		Creator: ctxUser.Username,
	}

	// 更新角色
	err = role.Update()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("更新角色失败: %s", err.Error())))
		return
	}

	// 如果更新成功，且更新了角色的keyword, 则更新casbin中policy
	if req.Keyword != roles[0].Keyword {
		// 获取policy
		rolePolicies, err := global.CasbinEnforcer.GetFilteredPolicy(0, roles[0].Keyword)
		if err != nil {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("更新角色失败: %s", err.Error())))
			return
		}
		if len(rolePolicies) == 0 {
			helper.Success(c, nil)
			return
		}
		rolePoliciesCopy := make([][]string, 0)
		// 替换keyword
		for _, policy := range rolePolicies {
			policyCopy := make([]string, len(policy))
			copy(policyCopy, policy)
			rolePoliciesCopy = append(rolePoliciesCopy, policyCopy)
			policy[0] = req.Keyword
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
			helper.ErrV2(c, helper.NewOperationError(fmt.Errorf("更新角色成功，但角色关键字关联的权限接口更新失败")))
			return
		}
		isRemoved, _ := global.CasbinEnforcer.RemovePolicies(rolePoliciesCopy)
		if !isRemoved {
			helper.ErrV2(c, helper.NewOperationError(fmt.Errorf("更新角色成功，但角色关键字关联的权限接口更新失败")))
			return
		}
		err = global.CasbinEnforcer.LoadPolicy()
		if err != nil {
			helper.ErrV2(c, helper.NewOperationError(fmt.Errorf("更新角色成功，但角色关键字关联角色的权限接口策略加载失败")))
			return
		}

	}

	// 更新角色成功处理用户信息缓存有两种做法:（这里使用第二种方法，因为一个角色下用户数量可能很多，第二种方法可以分散数据库压力）
	// 1.可以帮助用户更新拥有该角色的用户信息缓存,使用下面方法
	// err = ur.UpdateUserInfoCacheByRoleId(uint(roleId))
	// 2.直接清理缓存，让活跃的用户自己重新缓存最新用户信息
	model.ClearUserInfoCache()

	helper.Success(c, nil)
}

// RoleDeleteReq 删除资源结构体
type RoleDeleteReq struct {
	RoleIds []uint `json:"roleIds" validate:"required"`
}

// Delete 删除记录
func RoleDelete(c *gin.Context) {
	var req RoleDeleteReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取当前登陆用户最高等级角色
	minSort, _, err := current.GetCurrentUserMinRoleSort(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前用户最高角色等级失败: %s", err.Error())))
		return
	}

	// 获取角色信息
	roles := model.NewRoles()
	err = roles.GetRolesByIds(req.RoleIds)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取角色信息失败: %s", err.Error())))
		return
	}
	if len(roles) == 0 {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("未能获取到角色信息")))
		return
	}

	// 不能删除比自己角色等级高或相等的角色
	for _, role := range roles {
		if minSort >= role.Sort {
			helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("不能删除比自己角色等级高或相等的角色")))
			return
		}
	}

	// 删除角色
	err = roles.Delete()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("删除角色失败: %s", err.Error())))
		return
	}

	// 删除角色成功直接清理缓存，让活跃的用户自己重新缓存最新用户信息
	model.ClearUserInfoCache()
	helper.Success(c, nil)
}

// RoleGetMenuListReq 获取角色菜单列表结构体
type RoleGetMenuListReq struct {
	RoleID uint `json:"roleId" form:"roleId" validate:"required"`
}

// RoleGetMenuList 获取角色菜单列表
func RoleGetMenuList(c *gin.Context) {
	var req RoleGetMenuListReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	menus, err := model.GetRoleMenusById(req.RoleID)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取角色的权限菜单失败: "+err.Error())))
	}
	helper.Success(c, menus)

}

// RoleGetApiListReq 获取角色接口列表结构体
type RoleGetApiListReq struct {
	RoleID uint `json:"roleId" form:"roleId" validate:"required"`
}

// RoleGetApiList 获取角色接口列表
func RoleGetApiList(c *gin.Context) {
	var req RoleGetApiListReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	role := new(model.Role)
	err = role.Find(map[string]interface{}{"id": req.RoleID})
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取资源失败: "+err.Error())))
		return
	}

	policies, err := global.CasbinEnforcer.GetFilteredPolicy(0, role.Keyword)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取角色权限接口失败: %s", err.Error())))
		return
	}

	apis, err := apiMgrModel.ListAll()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取资源列表失败: "+err.Error())))
		return
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

	helper.Success(c, accessApis)
}

// RoleUpdateMenusReq 更新角色菜单结构体
type RoleUpdateMenusReq struct {
	RoleID  uint   `json:"roleId" validate:"required"`
	MenuIds []uint `json:"menuIds" validate:"required"`
}

// RoleUpdateMenus 更新角色菜单
func RoleUpdateMenus(c *gin.Context) {
	var req RoleUpdateMenusReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	roles := model.NewRoles()
	err = roles.GetRolesByIds([]uint{req.RoleID})
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取角色信息失败: %s", err.Error())))
		return
	}
	if len(roles) == 0 {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("未获取到角色信息")))
		return
	}

	// 当前用户角色排序最小值（最高等级角色）以及当前用户
	minSort, ctxUser, err := current.GetCurrentUserMinRoleSort(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前用户最高角色等级失败: %s", err.Error())))
		return
	}

	// TODO: roles[0] ?????
	global.Log.Infof("所有角色列表: %+v", roles)
	// (非管理员)不能更新比自己角色等级高或相等角色的权限菜单
	if minSort != 1 {
		if minSort >= roles[0].Sort {
			helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("不能更新比自己角色等级高或相等角色的权限菜单")))
			return
		}
	}

	// 获取当前用户所拥有的权限菜单
	ctxUserMenus, err := model.GetUserMenusByUserId(ctxUser.ID)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前用户的可访问菜单列表失败: "+err.Error())))
		return
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
		for _, id := range req.MenuIds {
			if !funk.Contains(ctxUserMenusIds, id) {
				helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("无权设置ID为%d的菜单", id)))
				return
			}
		}

		for _, id := range req.MenuIds {
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
			helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("获取菜单列表失败: "+err.Error())))
			return
		}
		for _, menuId := range req.MenuIds {
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
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("更新角色的权限菜单失败: "+err.Error())))
		return
	}

	helper.Success(c, nil)
}

// RoleUpdateApisReq 更新角色接口结构体
type RoleUpdateApisReq struct {
	RoleID uint   `json:"roleId" validate:"required"`
	ApiIds []uint `json:"apiIds" validate:"required"`
}

// RoleUpdateApis 更新角色接口
func RoleUpdateApis(c *gin.Context) {
	var req RoleUpdateApisReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 根据path中的角色ID获取该角色信息
	roles := model.NewRoles()
	if err := roles.GetRolesByIds([]uint{req.RoleID}); err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取角色信息失败: "+err.Error())))
		return
	}
	if len(roles) == 0 {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("未获取到角色信息")))
		return
	}

	// 当前用户角色排序最小值（最高等级角色）以及当前用户
	minSort, ctxUser, err := current.GetCurrentUserMinRoleSort(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前用户最高角色等级失败: %s", err.Error())))
		return
	}

	// (非管理员)不能更新比自己角色等级高或相等角色的权限菜单
	if minSort != 1 {
		if minSort >= roles[0].Sort {
			helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("不能更新比自己角色等级高或相等角色的权限菜单")))
			return
		}
	}

	// 获取当前用户所拥有的权限接口
	ctxRoles := ctxUser.Roles
	ctxRolesPolicies := make([][]string, 0)
	for _, role := range ctxRoles {
		policy, err := global.CasbinEnforcer.GetFilteredPolicy(0, role.Keyword)
		if err != nil {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前用户的可访问接口列表失败: "+err.Error())))
			return
		}
		ctxRolesPolicies = append(ctxRolesPolicies, policy...)
	}
	// 得到path中的角色ID对应角色能够设置的权限接口集合
	for _, policy := range ctxRolesPolicies {
		policy[0] = roles[0].Keyword
	}

	// 根据apiID获取接口详情
	apis, err := apiMgrModel.GetApisById(req.ApiIds)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("根据接口ID获取接口信息失败")))
		return
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
				helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("无权设置路径为%s,请求方式为%s的接口", reqPolicy[1], reqPolicy[2])))
				return
			}
		}
	}

	// 更新角色的权限接口
	err = model.UpdateRoleApis(roles[0].Keyword, reqRolePolicies)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("更新角色的权限接口失败")))
		return
	}
	helper.Success(c, nil)
}
