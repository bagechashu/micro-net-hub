package fieldrelation

import (
	"fmt"

	"micro-net-hub/internal/module/goldap/field_relation/model"
	"micro-net-hub/internal/server/helper"
	"micro-net-hub/internal/tools"

	"gorm.io/datatypes"

	"github.com/gin-gonic/gin"
)

// List 记录列表
func List(c *gin.Context) {
	// 获取数据列表
	frs, err := model.List()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("字段动态关系: %s", err.Error())))
		return
	}

	helper.Success(c, frs)
}

// FieldRelationAddReq 添加资源结构体
type FieldRelationAddReq struct {
	Flag       string            `json:"flag" validate:"required,min=1,max=20"`
	Attributes map[string]string `json:"attributes" validate:"required,gt=0"`
}

// Add 新建记录
func Add(c *gin.Context) {
	var req FieldRelationAddReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	if model.Exist(tools.H{"flag": req.Flag}) {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("对应平台的动态字段关系已存在，请勿重复添加")))
		return
	}

	attr, err := tools.MapToJson(req.Attributes)
	if err != nil {
		helper.ErrV2(c, helper.NewOperationError(fmt.Errorf("将map转成json失败: %s", err.Error())))
		return
	}

	frObj := model.FieldRelation{
		Flag:       req.Flag,
		Attributes: datatypes.JSON(attr),
	}

	// 创建接口
	err = model.Add(&frObj)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("创建动态字段关系失败: %s", err.Error())))
		return
	}

	helper.Success(c, nil)
}

// FieldRelationUpdateReq 更新资源结构体
type FieldRelationUpdateReq struct {
	ID         uint              `json:"id" validate:"required"`
	Flag       string            `json:"flag" validate:"required,min=1,max=20"`
	Attributes map[string]string `json:"attributes" validate:"required,gt=0"`
}

// Update 更新记录
func Update(c *gin.Context) {
	var req FieldRelationUpdateReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	filter := tools.H{"flag": req.Flag}

	if !model.Exist(filter) {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("对应平台的动态字段关系不存在")))
		return
	}

	oldData := new(model.FieldRelation)
	err = model.Find(filter, oldData)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(err))
		return
	}

	attr, err := tools.MapToJson(req.Attributes)
	if err != nil {
		helper.ErrV2(c, helper.NewOperationError(fmt.Errorf("将map转成json失败: %s", err.Error())))
		return
	}

	frObj := model.FieldRelation{
		Model:      oldData.Model,
		Flag:       req.Flag,
		Attributes: datatypes.JSON(attr),
	}

	err = model.Update(&frObj)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("更新动态字段关系失败: %s", err.Error())))
		return
	}

	helper.Success(c, nil)
}

// FieldRelationDeleteReq 删除资源结构体
type FieldRelationDeleteReq struct {
	FieldRelationIds []uint `json:"fieldRelationIds" validate:"required"`
}

// Delete 删除记录
func Delete(c *gin.Context) {
	var req FieldRelationDeleteReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	for _, id := range req.FieldRelationIds {
		filter := tools.H{"id": int(id)}
		if !model.Exist(filter) {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("动态字段关系不存在")))
			return
		}
	}
	// 删除
	err = model.Delete(req.FieldRelationIds)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("删除动态字段关系失败: %s", err.Error())))
		return
	}

	helper.Success(c, nil)
}
