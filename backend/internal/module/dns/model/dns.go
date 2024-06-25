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
	Name       string       `gorm:"type:varchar(64)" json:"name"`
	DnsRecords []*DnsRecord `gorm:"foreignKey:ZoneID;references:ID" json:"records"`
	Creator    string       `gorm:"type:varchar(20);comment:'创建人'" json:"creator" form:"creator"`
}

func (dz *DnsZone) Find(filter map[string]interface{}) error {
	if err := global.DB.Table("dns_zones").Where(filter).First(&dz).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 DnsZone 失败: %w", err))
	}
	return nil
}

func (dz *DnsZone) Add() error {
	if err := global.DB.Create(&dz).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("添加 DnsZone 失败: %w", err))
	}
	return nil
}

func (dz *DnsZone) Update() error {
	if err := global.DB.Save(&dz).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("更新 DnsZone 失败: %w", err))
	}
	return nil
}

func (dz *DnsZone) Delete() error {
	if err := global.DB.Unscoped().Delete(&dz).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("删除 DnsZone 失败: %w", err))
	}
	return nil
}

func (dz *DnsZone) DeleteWithRecords() error {
	if err := global.DB.Unscoped().Delete(&dz).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("删除 DnsZone 失败: %w", err))
	}
	if err := global.DB.Where("zone_id = ?", dz.ID).Delete(&DnsRecord{}).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("删除 DnsRecord 失败: %w", err))
	}
	return nil
}

type DnsZones []*DnsZone

var DnsZonesCache = cache.New(24*time.Hour, 48*time.Hour)

func cacheDnsZonesKeygen() string {
	return "ldz_all"
}

func CacheDnsZonesGet() (dnsZones DnsZones) {
	dzCacheKey := cacheDnsZonesKeygen()
	if cacheDzs, found := DnsZonesCache.Get(dzCacheKey); found {
		return cacheDzs.(DnsZones)
	}
	return nil
}

func CacheDnsZonesSet(dnsZones DnsZones) {
	dzCacheKey := cacheDnsZonesKeygen()
	DnsZonesCache.Set(dzCacheKey, dnsZones, cache.DefaultExpiration)
}

func CacheDnsZonesDel() {
	dzCacheKey := cacheDnsZonesKeygen()
	DnsZonesCache.Delete(dzCacheKey)
}

func CacheDnsZonesClear() {
	DnsZonesCache.Flush()
}

func (zrs *DnsZones) FindAll() error {
	if err := global.DB.Find(&zrs).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 DnsZone 失败: %w", err))
	}
	return nil
}

func (dzs *DnsZones) FindWithRecords() error {
	if err := global.DB.Preload("DnsRecords").Find(&dzs).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 DnsZone 失败: %w", err))
	}
	return nil
}

type DnsRecord struct {
	gorm.Model
	ZoneID  uint   `gorm:"type:uint" json:"zone_id"`
	Type    string `gorm:"type:varchar(64)" json:"type"`
	Host    string `gorm:"type:varchar(64)" json:"host"`
	Value   string `gorm:"type:varchar(64)" json:"value"`
	Ttl     uint32 `gorm:"type:uint" json:"ttl"`
	Creator string `gorm:"type:varchar(20);comment:'创建人'" json:"creator" form:"creator"`
}

func (dr *DnsRecord) Find(filter map[string]interface{}) error {
	if err := global.DB.Table("dns_records").Where(filter).First(&dr).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 DnsZone 失败: %w", err))
	}
	return nil
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

type DnsRecords []*DnsRecord

var DnsRecordsCache = cache.New(24*time.Hour, 48*time.Hour)

func cacheDnsRecordKeygen(zone_id uint, host string) string {
	return fmt.Sprintf("ldr_%d_%s", zone_id, host)
}
func CacheDnsRecordGet(zone_id uint, host string) (dnsRecords DnsRecords) {
	drCacheKey := cacheDnsRecordKeygen(zone_id, host)
	if cacheDrs, found := DnsRecordsCache.Get(drCacheKey); found {
		return cacheDrs.(DnsRecords)
	}
	return nil
}

func CacheDnsRecordSet(zone_id uint, host string, dnsRecords DnsRecords) {
	drCacheKey := cacheDnsRecordKeygen(zone_id, host)
	DnsRecordsCache.Set(drCacheKey, dnsRecords, cache.DefaultExpiration)
}

func CacheDnsRecordDel(zone_id uint, host string) {
	drCacheKey := cacheDnsRecordKeygen(zone_id, host)
	if _, found := DnsRecordsCache.Get(drCacheKey); found {
		DnsRecordsCache.Delete(drCacheKey)
	}
}

func CacheDnsRecordsClear() {
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
