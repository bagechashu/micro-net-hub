package model

import (
	"fmt"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/server/helper"
	"micro-net-hub/internal/tools"

	"gorm.io/gorm"
)

// 导航分组
type NavGroup struct {
	gorm.Model
	Title    string     `gorm:"type:varchar(20);comment:'网址分组标题'" json:"title"`
	Name     string     `gorm:"type:varchar(20);unique;comment:'网址分组名'" json:"name"`
	NavSites []*NavSite `json:"sites"`
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

func (g *NavGroup) DeleteWithSites() error {
	if err := global.DB.Where("name = ?", g.Name).First(&g).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 NavGroup 失败: " + err.Error()))
	}
	for _, ni := range g.NavSites {
		if err := global.DB.Delete(&ni).Error; err != nil {
			return helper.NewMySqlError(fmt.Errorf("删除 NavSite 失败: " + err.Error()))
		}
	}
	if err := global.DB.Delete(&g).Error; err != nil {
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

func (gs *NavGroups) Find(req *NavReq) (err error) {
	pageReq := tools.NewPageOption(req.PageNum, req.PageSize)
	err = global.DB.Model(&NavGroup{}).Offset(pageReq.PageNum).Limit(pageReq.PageSize).Find(&gs).Error

	return err
}

func (gs *NavGroups) FindWithSites() error {
	if err := global.DB.Debug().Preload("NavSites").Find(&gs).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 NavGroups 失败: " + err.Error()))
	}
	return nil
}

type NavSite struct {
	gorm.Model
	NavGroupID  uint   `gorm:"type:uint;comment:'网址分组ID'" json:"groupid"`
	IconUrl     string `gorm:"type:varchar(100);comment:'网址Icon'" json:"icon"`
	Name        string `gorm:"type:varchar(20);unique;comment:'网址名'" json:"name"`
	Description string `gorm:"type:varchar(50);comment:'网址描述'" json:"desc"`
	Link        string `gorm:"type:varchar(100);comment:'网址链接'" json:"link"`
	DocUrl      string `gorm:"type:varchar(100);comment:'网址文档'" json:"doc,omitempty"`
}

func (ni *NavSite) FindByName(name string) error {
	if err := global.DB.Where("name = ?", name).First(&ni).Error; err != nil {
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
	if err := global.DB.Delete(&ni).Error; err != nil {
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

func (is *NavSites) Find(req *NavReq) (err error) {
	pageReq := tools.NewPageOption(req.PageNum, req.PageSize)
	err = global.DB.Model(&NavSite{}).Offset(pageReq.PageNum).Limit(pageReq.PageSize).Find(&is).Error

	return err
}

type NavReq struct {
	PageNum  int `json:"pageNum" form:"pageNum"`
	PageSize int `json:"pageSize" form:"pageSize"`
}

type NavRsp struct {
	Groups     []*NavGroup `json:"groups"`
	Sites      []*NavSite  `json:"sites"`
	GroupTotal int64       `json:"groupTotal"`
	SiteTotal  int64       `json:"siteTotal"`
}
