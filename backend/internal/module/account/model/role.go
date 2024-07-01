package model

import (
	"errors"
	"fmt"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/tools"
	"strings"

	"gorm.io/gorm"
)

type RoleStatus uint8

const (
	RoleNormal   RoleStatus = iota + 1 // 1 正常
	RoleDisabled                       // 2 禁用
)

const SuperAdminRoleID = 1

type Role struct {
	gorm.Model
	Name    string     `gorm:"type:varchar(20);not null;unique" json:"name"`
	Keyword string     `gorm:"type:varchar(20);not null;unique" json:"keyword"`
	Remark  string     `gorm:"type:varchar(100);comment:'备注'" json:"remark"`
	Status  RoleStatus `gorm:"type:tinyint(1);default:1;comment:'1正常, 2禁用'" json:"status"`
	Sort    uint       `gorm:"type:int(3);default:999;comment:'角色排序(排序越大权限越低, 不能查看比自己序号小的角色, 不能编辑同序号用户权限, 排序为1表示超级管理员)'" json:"sort"`
	Creator string     `gorm:"type:varchar(20);" json:"creator"`
	Users   []*User    `gorm:"many2many:user_roles" json:"users"`
	Menus   []*Menu    `gorm:"many2many:role_menus;" json:"menus"` // 角色菜单多对多关系
}

// Exist 判断资源是否存在
func (r *Role) Exist(filter map[string]interface{}) bool {
	err := global.DB.Order("created_at DESC").Where(filter).First(&r).Error
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

// Add 创建资源
func (r *Role) Add() error {
	return global.DB.Create(&r).Error
}

// Update 更新资源
func (r *Role) Update() error {
	return global.DB.Model(&Role{}).Where("id = ?", r.ID).Updates(&r).Error
}

// UpdateRoleMenus 更新角色的权限菜单
func (r *Role) UpdateRoleMenus() error {
	return global.DB.Model(r).Association("Menus").Replace(r.Menus)
}

// Find 获取单个资源
func (r *Role) Find(filter map[string]interface{}) error {
	return global.DB.Where(filter).First(&r).Error
}

// Count 获取资源总数
func RoleCount() (int64, error) {
	var count int64
	err := global.DB.Model(&Role{}).Count(&count).Error
	return count, err
}

// GetRoleUsers 获取角色下的用户
func GetRoleUsersByKeyword(keyword string) ([]*User, error) {
	var role Role
	err := global.DB.Where("keyword = ?", keyword).Preload("Users").First(&role).Error
	return role.Users, err
}

// GetRoleMenusById 获取角色的权限菜单
func GetRoleMenusById(roleId uint) ([]*Menu, error) {
	var role Role
	err := global.DB.Where("id = ?", roleId).Preload("Menus").First(&role).Error
	return role.Menus, err
}

// UpdateRoleApis 更新角色的权限接口（先全部删除再新增）
func UpdateRoleApis(roleKeyword string, reqRolePolicies [][]string) error {
	// 先获取path中的角色ID对应角色已有的police(需要先删除的)
	err := global.CasbinEnforcer.LoadPolicy()
	if err != nil {
		return errors.New("角色的权限接口策略加载失败")
	}
	rmPolicies, err := global.CasbinEnforcer.GetFilteredPolicy(0, roleKeyword)
	if err != nil {
		return errors.New("角色的权限接口策略加载失败")
	}
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

type Roles []*Role

func NewRoles() Roles {
	return make([]*Role, 0)
}

// List 获取数据列表
func (rs *Roles) List(req *Role, pageNum int, pageSize int) error {
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

	pageReq := tools.NewPageOption(pageNum, pageSize)
	err := db.Offset(pageReq.PageNum).Limit(pageReq.PageSize).Find(&rs).Error
	return err
}

// Delete 删除资源
func (rs *Roles) Delete() error {
	for _, r := range *rs {
		if err := global.DB.Unscoped().Delete(&r).Error; err != nil {
			return err
		}

		// 删除成功就删除casbin policy
		roleKeyword := r.Keyword
		policies, err := global.CasbinEnforcer.GetFilteredPolicy(0, roleKeyword)
		if err != nil {
			return errors.New("角色的权限接口策略加载失败")
		}
		if len(policies) > 0 {
			isRemoved, _ := global.CasbinEnforcer.RemovePolicies(policies)
			if !isRemoved {
				return fmt.Errorf("角色 %s 删除成功, 角色关联权限 %+v 删除失败", r.Keyword, policies)
			}
		}
	}

	return nil
}

// 根据角色ID获取角色
func (rs *Roles) GetRolesByIds(roleIds []uint) error {
	err := global.DB.Where("id IN (?)", roleIds).Find(&rs).Error
	return err
}
