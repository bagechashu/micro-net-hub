package model

import (
	"errors"
	"micro-net-hub/internal/global"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type FieldRelation struct {
	gorm.Model
	Flag       string
	Attributes datatypes.JSON
}

// Exist 判断资源是否存在
func Exist(filter map[string]interface{}) bool {
	var dataObj FieldRelation
	err := global.DB.Order("created_at DESC").Where(filter).First(&dataObj).Error
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

// Count 获取资源总数
func Count() (int64, error) {
	var count int64
	err := global.DB.Model(&FieldRelation{}).Count(&count).Error
	return count, err
}

// Add 创建资源
func Add(fieldRelation *FieldRelation) error {
	return global.DB.Create(fieldRelation).Error
}

// Update 更新资源
func Update(fieldRelation *FieldRelation) error {
	return global.DB.Model(&FieldRelation{}).Where("id = ?", fieldRelation.ID).Updates(fieldRelation).Error
}

// Find 获取单个资源
func Find(filter map[string]interface{}, data *FieldRelation) error {
	return global.DB.Where(filter).First(&data).Error
}

// List 获取数据列表
func List() (fieldRelations []*FieldRelation, err error) {
	err = global.DB.Find(&fieldRelations).Error
	return fieldRelations, err
}

// 批量删除资源
func Delete(fieldRelationIds []uint) error {
	return global.DB.Where("id IN (?)", fieldRelationIds).Unscoped().Delete(&FieldRelation{}).Error
}
