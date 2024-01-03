package model

import (
	"errors"

	"micro-net-hub/internal/global"

	"github.com/thoas/go-funk"
	"gorm.io/gorm"
)

type MenuService struct{}

// Exist 判断资源是否存在
func (s MenuService) Exist(filter map[string]interface{}) bool {
	var dataObj Menu
	err := global.DB.Debug().Order("created_at DESC").Where(filter).First(&dataObj).Error
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

// Count 获取资源总数
func (s MenuService) Count() (int64, error) {
	var count int64
	err := global.DB.Model(&Menu{}).Count(&count).Error
	return count, err
}

// Add 创建资源
func (s MenuService) Add(menu *Menu) error {
	return global.DB.Create(menu).Error
}

// Update 更新资源
func (s MenuService) Update(menu *Menu) error {
	return global.DB.Model(&Menu{}).Where("id = ?", menu.ID).Updates(menu).Error
}

// Find 获取单个资源
func (s MenuService) Find(filter map[string]interface{}, data *Menu) error {
	return global.DB.Where(filter).First(&data).Error
}

// List 获取数据列表
func (s MenuService) List() (menus []*Menu, err error) {
	err = global.DB.Order("sort").Find(&menus).Error
	return menus, err
}

// List 获取数据列表
func (s MenuService) ListUserMenus(roleIds []uint) (menus []*Menu, err error) {
	err = global.DB.Where("id IN (select menu_id as id from role_menus where role_id IN (?))", roleIds).Order("sort").Find(&menus).Error
	return menus, err
}

// 批量删除资源
func (s MenuService) Delete(menuIds []uint) error {
	return global.DB.Where("id IN (?)", menuIds).Select("Roles").Unscoped().Delete(&Menu{}).Error
}

// GetUserMenusByUserId 根据用户ID获取用户的权限(可访问)菜单列表
func (s MenuService) GetUserMenusByUserId(userId uint) ([]*Menu, error) {
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

// // GetMenuTree 获取菜单树
// func (s MenuService) GetMenuTree() ([]*Menu, error) {
// 	var menus []*Menu
// 	err := common.DB.Order("sort").Find(&menus).Error
// 	// parentId为0的是根菜单
// 	return GenMenuTree(0, menus), err
// }
