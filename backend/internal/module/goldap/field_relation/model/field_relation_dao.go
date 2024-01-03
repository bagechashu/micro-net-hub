package model

import (
	"errors"

	"micro-net-hub/internal/global"

	"gorm.io/gorm"
)

type FieldRelationService struct{}

// Exist 判断资源是否存在
func (s FieldRelationService) Exist(filter map[string]interface{}) bool {
	var dataObj FieldRelation
	err := global.DB.Debug().Order("created_at DESC").Where(filter).First(&dataObj).Error
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

// Count 获取资源总数
func (s FieldRelationService) Count() (int64, error) {
	var count int64
	err := global.DB.Model(&FieldRelation{}).Count(&count).Error
	return count, err
}

// Add 创建资源
func (s FieldRelationService) Add(fieldRelation *FieldRelation) error {
	return global.DB.Create(fieldRelation).Error
}

// Update 更新资源
func (s FieldRelationService) Update(fieldRelation *FieldRelation) error {
	return global.DB.Model(&FieldRelation{}).Where("id = ?", fieldRelation.ID).Updates(fieldRelation).Error
}

// Find 获取单个资源
func (s FieldRelationService) Find(filter map[string]interface{}, data *FieldRelation) error {
	return global.DB.Where(filter).First(&data).Error
}

// List 获取数据列表
func (s FieldRelationService) List() (fieldRelations []*FieldRelation, err error) {
	err = global.DB.Find(&fieldRelations).Error
	return fieldRelations, err
}

// 批量删除资源
func (s FieldRelationService) Delete(fieldRelationIds []uint) error {
	return global.DB.Where("id IN (?)", fieldRelationIds).Unscoped().Delete(&FieldRelation{}).Error
}
