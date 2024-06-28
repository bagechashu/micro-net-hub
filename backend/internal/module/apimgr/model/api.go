package model

import (
	"errors"
	"fmt"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/tools"
	"strings"

	"gorm.io/gorm"
)

type Api struct {
	gorm.Model
	Method   string `gorm:"type:varchar(20);comment:'请求方式'" json:"method"`
	Path     string `gorm:"type:varchar(100);comment:'访问路径'" json:"path"`
	Category string `gorm:"type:varchar(50);comment:'所属类别'" json:"category"`
	Remark   string `gorm:"type:varchar(100);comment:'备注'" json:"remark"`
	Creator  string `gorm:"type:varchar(20);comment:'创建人'" json:"creator"`
}

// List 获取数据列表
func List(api *Api, pageNum int, pageSize int) ([]*Api, error) {
	var list []*Api
	db := global.DB.Model(&Api{}).Order("created_at DESC")

	method := strings.TrimSpace(api.Method)
	if method != "" {
		db = db.Where("method LIKE ?", fmt.Sprintf("%%%s%%", method))
	}
	path := strings.TrimSpace(api.Path)
	if path != "" {
		db = db.Where("path LIKE ?", fmt.Sprintf("%%%s%%", path))
	}
	category := strings.TrimSpace(api.Category)
	if category != "" {
		db = db.Where("category LIKE ?", fmt.Sprintf("%%%s%%", category))
	}
	creator := strings.TrimSpace(api.Creator)
	if creator != "" {
		db = db.Where("creator LIKE ?", fmt.Sprintf("%%%s%%", creator))
	}

	pageReq := tools.NewPageOption(pageNum, pageSize)
	err := db.Offset(pageReq.PageNum).Limit(pageReq.PageSize).Find(&list).Error
	return list, err
}

// List 获取数据列表
func ListAll() (list []*Api, err error) {
	err = global.DB.Model(&Api{}).Order("created_at DESC").Find(&list).Error

	return list, err
}

// Count 获取数据总数
func Count() (int64, error) {
	var count int64
	err := global.DB.Model(&Api{}).Count(&count).Error
	return count, err
}

// Add 添加资源
func Add(api *Api) error {
	return global.DB.Create(api).Error
}

// Update 更新资源
func Update(api *Api) error {
	// 根据id获取接口信息
	var oldApi Api
	err := global.DB.First(&oldApi, api.ID).Error
	if err != nil {
		return errors.New("根据接口ID获取接口信息失败")
	}
	err = global.DB.Model(api).Where("id = ?", api.ID).Updates(api).Error
	if err != nil {
		return err
	}
	// 更新了method和path就更新casbin中policy
	if oldApi.Path != api.Path || oldApi.Method != api.Method {
		policies, err := global.CasbinEnforcer.GetFilteredPolicy(1, oldApi.Path, oldApi.Method)
		if err != nil {
			return errors.New("更新权限接口失败")
		}
		// 接口在casbin的policy中存在才进行操作
		if len(policies) > 0 {
			// 先删除
			isRemoved, _ := global.CasbinEnforcer.RemovePolicies(policies)
			if !isRemoved {
				return errors.New("更新权限接口失败")
			}
			for _, policy := range policies {
				policy[1] = api.Path
				policy[2] = api.Method
			}
			// 新增
			isAdded, _ := global.CasbinEnforcer.AddPolicies(policies)
			if !isAdded {
				return errors.New("更新权限接口失败")
			}
			// 加载policy
			err := global.CasbinEnforcer.LoadPolicy()
			if err != nil {
				return errors.New("更新权限接口成功，权限接口策略加载失败")
			} else {
				return err
			}
		}
	}
	return err
}

// Find 获取单个资源
func Find(filter map[string]interface{}, data *Api) error {
	return global.DB.Where(filter).First(&data).Error
}

// Exist 判断资源是否存在
func Exist(filter map[string]interface{}) bool {
	var dataObj Api
	err := global.DB.Order("created_at DESC").Where(filter).First(&dataObj).Error
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

// Delete 批量删除
func Delete(ids []uint) error {
	var apis []Api
	for _, id := range ids {
		// 根据ID获取用户
		api := new(Api)
		err := Find(map[string]interface{}{"id": id}, api)
		if err != nil {
			return fmt.Errorf("根据ID获取接口信息失败: %v", err)
		}
		apis = append(apis, *api)
	}

	err := global.DB.Where("id IN (?)", ids).Unscoped().Delete(&Api{}).Error
	// 如果删除成功，删除casbin中policy
	if err == nil {
		for _, api := range apis {
			policies, err := global.CasbinEnforcer.GetFilteredPolicy(1, api.Path, api.Method)
			if err != nil {
				return errors.New("删除权限接口失败")
			}
			if len(policies) > 0 {
				isRemoved, _ := global.CasbinEnforcer.RemovePolicies(policies)
				if !isRemoved {
					return errors.New("删除权限接口失败")
				}
			}
		}
		// 重新加载策略
		err := global.CasbinEnforcer.LoadPolicy()
		if err != nil {
			return errors.New("删除权限接口成功，权限接口策略加载失败")
		} else {
			return err
		}
	}
	return err
}

// GetApisById 根据接口ID获取接口列表
func GetApisById(apiIds []uint) ([]*Api, error) {
	var apis []*Api
	err := global.DB.Where("id IN (?)", apiIds).Find(&apis).Error
	return apis, err
}
