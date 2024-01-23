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
	Title    string     `gorm:"type:varchar(20);not null;comment:'网址分组标题'" json:"title" validate:"required"`
	Name     string     `gorm:"type:varchar(20);not null;unique;comment:'网址分组名'" json:"name" validate:"required"`
	Creator  string     `gorm:"type:varchar(20);comment:'创建人'" json:"creator" form:"creator"`
	NavSites []*NavSite `json:"sites"`
}

type NavGroupAddReq struct {
	Title string `json:"title" validate:"required"`
	Name  string `json:"name" validate:"required"`
}

type NavGroupUpdateReq struct {
	Title string `json:"title" validate:"required"`
	Name  string `json:"name" validate:"required"`
}

type NavGroupDeleteReq struct {
	Ids []uint `json:"ids" validate:"required"`
}

func (g *NavGroup) FindByName(name string) error {
	if err := global.DB.Where("name = ?", name).First(&g).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 NavGroup 失败: " + err.Error()))
	}
	return nil
}

func (g *NavGroup) FindById(id uint) error {
	if err := global.DB.First(&g, id).Error; err != nil {
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
	if err := global.DB.Unscoped().Delete(&g).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("删除 NavGroup 失败: " + err.Error()))
	}
	return nil
}

func (g *NavGroup) DeleteWithSites() error {
	if err := global.DB.Where("name = ?", g.Name).Preload("NavSites").First(&g).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 NavGroup 失败: " + err.Error()))
	}
	for _, ni := range g.NavSites {
		if err := global.DB.Unscoped().Delete(&ni).Error; err != nil {
			return helper.NewMySqlError(fmt.Errorf("删除 NavSite 失败: " + err.Error()))
		}
	}
	if err := global.DB.Unscoped().Delete(&g).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("删除 NavGroup 失败: " + err.Error()))
	}
	return nil
}

func GroupCount() (count int64, err error) {
	err = global.DB.Model(&NavGroup{}).Count(&count).Error
	return count, err
}

type NavGroups []*NavGroup

func NewNavGroups() *NavGroups {
	return &NavGroups{}
}

func (gs *NavGroups) Find() (err error) {
	err = global.DB.Model(&NavGroup{}).Find(&gs).Error

	return err
}

func (gs *NavGroups) FindWithSites() error {
	if err := global.DB.Preload("NavSites").Find(&gs).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 NavGroups 失败: " + err.Error()))
	}
	return nil
}

type NavSite struct {
	gorm.Model
	Name        string `gorm:"type:varchar(20);not null;unique;comment:'网址名'" json:"name" validate:"required"`
	NavGroupID  uint   `gorm:"type:uint;not null;comment:'网址分组ID'" json:"groupid" validate:"required"`
	IconUrl     string `gorm:"type:varchar(100);comment:'网址Icon'" json:"icon"`
	Description string `gorm:"type:varchar(50);comment:'网址描述'" json:"desc"`
	Link        string `gorm:"type:varchar(100);not null;comment:'网址链接'" json:"link" validate:"required"`
	DocUrl      string `gorm:"type:varchar(100);comment:'网址文档'" json:"doc,omitempty"`
	Creator     string `gorm:"type:varchar(20);comment:'创建人'" json:"creator" form:"creator"`
}

type NavSiteAddReq struct {
	Name        string `json:"name" validate:"required"`
	NavGroupID  uint   `json:"groupid" validate:"required"`
	IconUrl     string `json:"icon"`
	Description string `json:"desc"`
	Link        string `json:"link" validate:"required"`
	DocUrl      string `json:"doc,omitempty"`
}

type NavSiteUpdateReq struct {
	Name        string `json:"name" validate:"required"`
	NavGroupID  uint   `json:"groupid" validate:"required"`
	IconUrl     string `json:"icon"`
	Description string `json:"desc"`
	Link        string `json:"link" validate:"required"`
	DocUrl      string `json:"doc,omitempty"`
}

type NavSiteDeleteReq struct {
	Ids []uint `json:"ids" validate:"required"`
}

func (ni *NavSite) FindByName(name string) error {
	if err := global.DB.Where("name = ?", name).First(&ni).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 NavSite 失败: " + err.Error()))
	}
	return nil
}

func (ni *NavSite) FindById(id uint) error {
	if err := global.DB.First(&ni, id).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 NavSite 失败: " + err.Error()))
	}
	return nil
}

func (ni *NavSite) Add() error {
	if err := global.DB.Create(&ni).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("添加 NavSite 失败: " + err.Error()))
	}
	return nil
}

func (ni *NavSite) Update() error {
	if err := global.DB.Save(&ni).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("更新 NavSite 失败: " + err.Error()))
	}
	return nil
}

func (ni *NavSite) Delete() error {
	if err := global.DB.Unscoped().Delete(&ni).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("删除 NavSite 失败: " + err.Error()))
	}
	return nil
}

func SiteCount() (count int64, err error) {
	err = global.DB.Model(&NavSite{}).Count(&count).Error
	return count, err
}

type NavSites []*NavSite

func NewNavSites() *NavSites {
	return &NavSites{}
}

func (is *NavSites) Find() (err error) {
	err = global.DB.Model(&NavSite{}).Find(&is).Error
	return err
}
