package model

import (
	"fmt"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/server/helper"
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

type DnsZone struct {
	gorm.Model
	Name string `gorm:"type:varchar(64)" json:"name"`
}

func (zr *DnsZone) Add() error {
	if err := global.DB.Create(&zr).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("添加 DnsZone 失败: %w", err))
	}
	return nil
}

func (zr *DnsZone) Update() error {
	if err := global.DB.Save(&zr).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("更新 DnsZone 失败: %w", err))
	}
	return nil
}

func (zr *DnsZone) Delete() error {
	if err := global.DB.Unscoped().Delete(&zr).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("删除 DnsZone 失败: %w", err))
	}
	return nil
}

type DnsZones []*DnsZone

var DnsZonesCache = cache.New(24*time.Hour, 48*time.Hour)

func ClearDnsZonesCache() {
	DnsZonesCache.Flush()
}

func (zrs *DnsZones) FindAll() error {
	if err := global.DB.Find(&zrs).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 DnsZone 失败: %w", err))
	}
	return nil
}

type DnsRecord struct {
	gorm.Model
	ZoneID uint   `gorm:"type:uint" json:"zone_id"`
	Type   string `gorm:"type:varchar(64)" json:"type"`
	Host   string `gorm:"type:varchar(64)" json:"host"`
	Value  string `gorm:"type:varchar(64)" json:"value"`
	Ttl    uint32 `gorm:"type:uint" json:"ttl"`
	Status uint8  `gorm:"type:uint" json:"status"`
}

func (dr *DnsRecord) Add() error {
	if err := global.DB.Create(&dr).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("添加 DnsRecord 失败: %w", err))
	}
	return nil
}

func (dr *DnsRecord) Update() error {
	if err := global.DB.Save(&dr).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("更新 DnsRecord 失败: %w", err))
	}
	return nil
}

func (dr *DnsRecord) Delete() error {
	if err := global.DB.Unscoped().Delete(&dr).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("删除 DnsRecord 失败: %w", err))
	}
	return nil
}

func (dr *DnsRecord) Enable() error {
	dr.Status = 1
	if err := global.DB.Save(&dr).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("启用 DnsRecord 失败: %w", err))
	}
	return nil
}

func (dr *DnsRecord) Disable() error {
	dr.Status = 0
	if err := global.DB.Save(&dr).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("禁用 DnsRecord 失败: %w", err))
	}
	return nil
}

type DnsRecords []*DnsRecord

var DnsRecordsCache = cache.New(24*time.Hour, 48*time.Hour)

func ClearDnsRecordsCache() {
	DnsRecordsCache.Flush()
}

func (drs *DnsRecords) FindAll() error {
	if err := global.DB.Find(&drs).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 DnsRecords 失败: %w", err))
	}
	return nil
}

func (drs *DnsRecords) Find(filter map[string]interface{}) error {
	if err := global.DB.Table("dns_records").Where(filter).Find(&drs).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 DnsRecord 失败: %w", err))
	}
	return nil
}
