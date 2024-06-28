package model

import (
	"errors"
	"micro-net-hub/internal/global"

	"github.com/thoas/go-funk"
	"gorm.io/gorm"
)

type MenuStatus uint

const (
	MenuEnabled  MenuStatus = iota + 1 // 菜单状态：正常，默认
	MenuDisabled                       // 菜单状态：禁用
)

type MenuVisibility uint

const (
	MenuHidden  MenuVisibility = iota + 1 // 侧边栏隐藏
	MenuVisible                           // 侧边栏显示，默认
)

type MenuCachePolicy uint

const (
	MenuNoCache   MenuCachePolicy = iota + 1 // 不被<keep-alive>缓存
	MenuWithCache                            // 被<keep-alive>缓存，默认
)

type MenuRootRouteDisplay uint

const (
	MenuAlwaysHide     MenuRootRouteDisplay = iota + 1 // 忽略规则，根路由一直隐藏
	MenuDefaultDisplay                                 // 根据规则显示根路由，默认
)

type MenuBreadcrumbVisibility uint

const (
	MenuVisibleBreadcrumb MenuBreadcrumbVisibility = iota + 1 // 面包屑显示，默认
	MenuHiddenBreadcrumb                                      // 面包屑隐藏
)

type Menu struct {
	gorm.Model
	Name       string                   `gorm:"type:varchar(50);comment:'菜单名称(英文名, 可用于国际化)'" json:"name"`
	Title      string                   `gorm:"type:varchar(50);comment:'菜单标题(无法国际化时使用)'" json:"title"`
	Icon       string                   `gorm:"type:varchar(50);comment:'菜单图标'" json:"icon"`
	Path       string                   `gorm:"type:varchar(100);comment:'菜单访问路径'" json:"path"`
	Redirect   string                   `gorm:"type:varchar(100);comment:'重定向路径'" json:"redirect"`
	Component  string                   `gorm:"type:varchar(100);comment:'前端组件路径'" json:"component"`
	Sort       uint                     `gorm:"type:int(3);default:999;comment:'菜单顺序(1-999)'" json:"sort"`
	Status     MenuStatus               `gorm:"type:tinyint(1);default:1;comment:'菜单状态(正常/禁用, 默认正常)'" json:"status"`
	Hidden     MenuVisibility           `gorm:"type:tinyint(1);default:2;comment:'菜单在侧边栏隐藏(1隐藏，2显示)'" json:"hidden"`
	NoCache    MenuCachePolicy          `gorm:"type:tinyint(1);default:2;comment:'菜单是否被 <keep-alive> 缓存(1不缓存，2缓存)'" json:"noCache"`
	AlwaysShow MenuRootRouteDisplay     `gorm:"type:tinyint(1);default:2;comment:'忽略之前定义的规则，一直显示根路由(1忽略，2不忽略)'" json:"alwaysShow"`
	Breadcrumb MenuBreadcrumbVisibility `gorm:"type:tinyint(1);default:1;comment:'面包屑可见性(可见/隐藏, 默认可见)'" json:"breadcrumb"`
	ActiveMenu string                   `gorm:"type:varchar(100);comment:'在其它路由时，想在侧边栏高亮的路由'" json:"activeMenu"`
	ParentId   uint                     `gorm:"default:0;comment:'父菜单编号(编号为0时表示根菜单)'" json:"parentId"`
	Creator    string                   `gorm:"type:varchar(20);comment:'创建人'" json:"creator"`
	Children   []*Menu                  `gorm:"-" json:"children"`                  // 子菜单集合
	Roles      []*Role                  `gorm:"many2many:role_menus;" json:"roles"` // 角色菜单多对多关系
}

// Exist 判断资源是否存在
func (m *Menu) Exist(filter map[string]interface{}) bool {
	err := global.DB.Order("created_at DESC").Where(filter).First(&m).Error
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

// Add 创建资源
func (m *Menu) Add() error {
	return global.DB.Create(m).Error
}

// Update 更新资源
func (m *Menu) Update() error {
	// https://gorm.io/zh_CN/docs/update.html
	// NOTE When updating with struct, GORM will only update non-zero fields. You might want to use map to update attributes or use Select to specify fields to update
	// global.Log.Infof("menu check ParentId before Update: %v", m.ParentId)
	if m.ParentId == 0 {
		err := global.DB.Model(&Menu{}).Where("id = ?", m.ID).Updates(map[string]interface{}{"ParentId": 0}).Error
		if err != nil {
			return err
		}
	}
	err := global.DB.Where("id = ?", m.ID).Model(&Menu{}).Updates(m).Error
	if err != nil {
		return err
	}
	return nil
}

// 批量删除资源
func (m *Menu) Delete() error {
	return global.DB.Where("id = ?", m.ID).Unscoped().Delete(&Menu{}).Error
}

// Find 获取单个资源
func (m *Menu) Find(filter map[string]interface{}) error {
	return global.DB.Where(filter).First(&m).Error
}

type Menus []*Menu

func NewMenus() Menus {
	return make([]*Menu, 0)
}

// List 获取数据列表
func (ms *Menus) List() (err error) {
	err = global.DB.Order("sort").Find(&ms).Error
	return err
}

// List 获取数据列表
func (ms *Menus) ListUserMenus(roleIds []uint) (err error) {
	err = global.DB.Where("id IN (select menu_id as id from role_menus where role_id IN (?))", roleIds).Order("sort").Find(&ms).Error
	return err
}

// 批量删除资源
func (ms *Menus) Delete(menuIds []uint) error {
	return global.DB.Where("id IN (?)", menuIds).Unscoped().Delete(&Menu{}).Error
}

// Count 获取资源总数
func MenuCount() (int64, error) {
	var count int64
	err := global.DB.Model(&Menu{}).Count(&count).Error
	return count, err
}

// GetUserMenusByUserId 根据用户ID获取用户的权限(可访问)菜单列表
func GetUserMenusByUserId(userId uint) ([]*Menu, error) {
	// 获取用户
	var user User
	err := global.DB.Where("id = ?", userId).Preload("Roles").First(&user).Error
	if err != nil {
		return nil, err
	}
	// 获取角色
	roles := user.Roles
	// 所有角色的菜单集合
	allRoleMenus := make([]*Menu, 0)
	for _, role := range roles {
		var userRole Role
		err := global.DB.Where("id = ?", role.ID).Preload("Menus").First(&userRole).Error
		if err != nil {
			return nil, err
		}
		// 获取角色的菜单
		menus := userRole.Menus
		allRoleMenus = append(allRoleMenus, menus...)
	}

	// 所有角色的菜单集合去重
	allRoleMenusId := make([]int, 0)
	for _, menu := range allRoleMenus {
		allRoleMenusId = append(allRoleMenusId, int(menu.ID))
	}
	allRoleMenusIdUniq := funk.UniqInt(allRoleMenusId)
	allRoleMenusUniq := make([]*Menu, 0)
	for _, id := range allRoleMenusIdUniq {
		for _, menu := range allRoleMenus {
			if id == int(menu.ID) {
				allRoleMenusUniq = append(allRoleMenusUniq, menu)
				break
			}
		}
	}

	// 获取状态status为1的菜单
	accessMenus := make([]*Menu, 0)
	for _, menu := range allRoleMenusUniq {
		if menu.Status == 1 {
			accessMenus = append(accessMenus, menu)
		}
	}

	return accessMenus, err
}

// GenMenuTree 生成菜单树
func GenMenuTree(parentId uint, menus []*Menu) []*Menu {
	tree := make([]*Menu, 0)

	for _, m := range menus {
		if m.ParentId == parentId {
			children := GenMenuTree(m.ID, menus)
			m.Children = children
			tree = append(tree, m)
		}
	}
	return tree
}
