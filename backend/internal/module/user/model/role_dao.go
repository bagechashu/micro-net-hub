package model

import (
	"errors"
	"fmt"
	"strings"

	"micro-net-hub/internal/global"
	"micro-net-hub/internal/tools"

	"gorm.io/gorm"
)

type RoleService struct{}

// Exist 判断资源是否存在
func (s RoleService) Exist(filter map[string]interface{}) bool {
	var dataObj Role
	err := global.DB.Debug().Order("created_at DESC").Where(filter).First(&dataObj).Error
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

// List 获取数据列表
func (s RoleService) List(req *RoleListReq) ([]*Role, error) {
	var list []*Role
	db := global.DB.Model(&Role{}).Order("created_at DESC")

	name := strings.TrimSpace(req.Name)
	if name != "" {
		db = db.Where("name LIKE ?", fmt.Sprintf("%%%s%%", name))
	}
	keyword := strings.TrimSpace(req.Keyword)
	if keyword != "" {
		db = db.Where("keyword LIKE ?", fmt.Sprintf("%%%s%%", keyword))
	}
	status := req.Status
	if status != 0 {
		db = db.Where("status = ?", status)
	}

	pageReq := tools.NewPageOption(req.PageNum, req.PageSize)
	err := db.Offset(pageReq.PageNum).Limit(pageReq.PageSize).Find(&list).Error
	return list, err
}

// Count 获取资源总数
func (s RoleService) Count() (int64, error) {
	var count int64
	err := global.DB.Model(&Role{}).Count(&count).Error
	return count, err
}

// Add 创建资源
func (s RoleService) Add(role *Role) error {
	return global.DB.Create(role).Error
}

// Update 更新资源
func (s RoleService) Update(role *Role) error {
	return global.DB.Model(&Role{}).Where("id = ?", role.ID).Updates(role).Error
}

// Find 获取单个资源
func (s RoleService) Find(filter map[string]interface{}, data *Role) error {
	return global.DB.Where(filter).First(&data).Error
}

// Delete 删除资源
func (s RoleService) Delete(roleIds []uint) error {
	var roles []*Role
	err := global.DB.Where("id IN (?)", roleIds).Find(&roles).Error
	if err != nil {
		return err
	}
	err = global.DB.Select("Users", "Menus").Unscoped().Delete(&roles).Error
	// 删除成功就删除casbin policy
	if err == nil {
		for _, role := range roles {
			roleKeyword := role.Keyword
			rmPolicies := global.CasbinEnforcer.GetFilteredPolicy(0, roleKeyword)
			if len(rmPolicies) > 0 {
				isRemoved, _ := global.CasbinEnforcer.RemovePolicies(rmPolicies)
				if !isRemoved {
					return errors.New("删除角色成功, 删除角色关联权限接口失败")
				}
			}
		}

	}
	return err
}

// Delete 根据角色ID获取角色
func (s RoleService) GetRolesByIds(roleIds []uint) ([]*Role, error) {
	var list []*Role
	err := global.DB.Where("id IN (?)", roleIds).Find(&list).Error
	return list, err
}

// GetRoleMenusById 获取角色的权限菜单
func (s RoleService) GetRoleMenusById(roleId uint) ([]*Menu, error) {
	var role Role
	err := global.DB.Where("id = ?", roleId).Preload("Menus").First(&role).Error
	return role.Menus, err
}

// UpdateRoleMenus 更新角色的权限菜单
func (s RoleService) UpdateRoleMenus(role *Role) error {
	return global.DB.Model(role).Association("Menus").Replace(role.Menus)
}

// UpdateRoleApis 更新角色的权限接口（先全部删除再新增）
func (s RoleService) UpdateRoleApis(roleKeyword string, reqRolePolicies [][]string) error {
	// 先获取path中的角色ID对应角色已有的police(需要先删除的)
	err := global.CasbinEnforcer.LoadPolicy()
	if err != nil {
		return errors.New("角色的权限接口策略加载失败")
	}
	rmPolicies := global.CasbinEnforcer.GetFilteredPolicy(0, roleKeyword)
	if len(rmPolicies) > 0 {
		isRemoved, _ := global.CasbinEnforcer.RemovePolicies(rmPolicies)
		if !isRemoved {
			return errors.New("更新角色的权限接口失败")
		}
	}
	isAdded, _ := global.CasbinEnforcer.AddPolicies(reqRolePolicies)
	if !isAdded {
		return errors.New("更新角色的权限接口失败")
	}
	err = global.CasbinEnforcer.LoadPolicy()
	if err != nil {
		return errors.New("更新角色的权限接口成功，角色的权限接口策略加载失败")
	} else {
		return err
	}
}
