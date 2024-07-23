package model

import (
	"fmt"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/server/helper"

	"gorm.io/gorm"
)

// TODO: add an Notice Board
type NoticeBoard struct {
	gorm.Model
	Level   uint   `gorm:"type:bigint(20);not null;comment:'公告级别'" json:"level" validate:"required"`
	Content string `gorm:"type:varchar(200);not null;comment:'公告内容'" json:"content" validate:"required"`
	Creator string `gorm:"type:varchar(20);comment:'创建人'" json:"creator" form:"creator"`
}

const (
	LevelInfo     uint = iota + 1 // 1
	LevelWarning                  // 2
	LevelCritical                 // 3
)

func (n NoticeBoard) LevelText() string {
	switch n.Level {
	case LevelInfo:
		return "Info"
	case LevelWarning:
		return "Warning"
	case LevelCritical:
		return "Critical:"
	default:
		return "Other"
	}
}

func (n NoticeBoard) Find(filter map[string]interface{}) error {
	if err := global.DB.Model(&NoticeBoard{}).Where(filter).First(&n).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 NoticeBoard 失败: %w", err))
	}
	return nil
}

func (n NoticeBoard) Add() error {
	if err := global.DB.Create(&n).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("添加 NoticeBoard 失败: %w", err))
	}
	return nil
}

func (n NoticeBoard) Update() error {
	if err := global.DB.Save(&n).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("更新 NoticeBoard 失败: %w", err))
	}
	return nil
}

func (n NoticeBoard) Delete() error {
	if err := global.DB.Unscoped().Delete(&n).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("删除 NoticeBoard 失败: %w", err))
	}
	return nil
}

type NoticeBoards []*NoticeBoard

func (ns *NoticeBoards) Find(filter map[string]interface{}) error {
	if err := global.DB.Model(&NoticeBoard{}).Where(filter).Find(&ns).Error; err != nil {
		return helper.NewMySqlError(fmt.Errorf("获取 DnsZone 失败: %w", err))
	}
	return nil
}
