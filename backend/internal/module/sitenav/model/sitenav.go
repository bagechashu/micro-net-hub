package model

import (
	"fmt"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/server/helper"

	"gorm.io/gorm"
)

// 导航分组
type NavGroup struct {
	gorm.Model
	Title    string     `gorm:"type:varchar(20);comment:'网址分组标题'" json:"title"`
	Name     string     `gorm:"type:varchar(20);unique;comment:'网址分组名'" json:"name"`
	NavItems []*NavItem `json:"nav"`
}

func (g *NavGroup) FindByName(name string) error {
	if err := global.DB.Where("name = ?", name).First(&g).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 NavGroup 失败: " + err.Error()))
	}
	return nil
}

func (g *NavGroup) Add() error {
	if err := global.DB.Create(&g).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("添加 NavGroup 失败: " + err.Error()))
	}
	return nil
}

func (g *NavGroup) Update() error {
	if err := global.DB.Save(&g).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("更新 NavGroup 失败: " + err.Error()))
	}
	return nil
}

func (g *NavGroup) Delete() error {
	if err := global.DB.Delete(&g).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("删除 NavGroup 失败: " + err.Error()))
	}
	return nil
}

func (g *NavGroup) DeleteWithItems() error {
	if err := global.DB.Where("name = ?", g.Name).First(&g).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 NavGroup 失败: " + err.Error()))
	}
	for _, ni := range g.NavItems {
		if err := global.DB.Delete(&ni).Error; err != nil {
			return helper.NewMySqlError(fmt.Errorf("删除 NavItem 失败: " + err.Error()))
		}
	}
	if err := global.DB.Delete(&g).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("删除 NavGroup 失败: " + err.Error()))
	}
	return nil
}

type NavGroups []NavGroup

func (gs *NavGroups) Find() error {
	if err := global.DB.Find(&gs).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 NavGroups 失败: " + err.Error()))
	}
	return nil
}

func (gs *NavGroups) FindWithItems() error {
	if err := global.DB.Preload("NavItems").Find(&gs).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 NavGroups 失败: " + err.Error()))
	}
	return nil
}

type NavItem struct {
	gorm.Model
	NavGroupID  uint   `gorm:"type:uint;comment:'网址分组ID'" json:"-"`
	IconUrl     string `gorm:"type:varchar(100);comment:'网址Icon'" json:"icon"`
	Name        string `gorm:"type:varchar(20);unique;comment:'网址名'" json:"name"`
	Description string `gorm:"type:varchar(50);comment:'网址描述'" json:"desc"`
	Link        string `gorm:"type:varchar(100);comment:'网址链接'" json:"link"`
	DocUrl      string `gorm:"type:varchar(100);comment:'网址文档'" json:"doc,omitempty"`
}

func (ni *NavItem) FindByName(name string) error {
	if err := global.DB.Where("name = ?", name).First(&ni).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 NavItem 失败: " + err.Error()))
	}
	return nil
}

func (ni *NavItem) Add() error {
	if err := global.DB.Create(&ni).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("添加 NavItem 失败: " + err.Error()))
	}
	return nil
}

func (ni *NavItem) Update() error {
	if err := global.DB.Save(&ni).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("更新 NavItem 失败: " + err.Error()))
	}
	return nil
}

func (ni *NavItem) Delete() error {
	if err := global.DB.Delete(&ni).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("删除 NavItem 失败: " + err.Error()))
	}
	return nil
}
