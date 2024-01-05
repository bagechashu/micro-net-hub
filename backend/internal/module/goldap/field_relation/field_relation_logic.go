package fieldrelation

import (
	"fmt"

	"micro-net-hub/internal/module/goldap/field_relation/model"
	"micro-net-hub/internal/tools"

	"gorm.io/datatypes"

	"github.com/gin-gonic/gin"
)

// Add 添加数据
func Add(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*model.FieldRelationAddReq)
	if !ok {
		return nil, tools.ReqAssertErr
	}
	_ = c

	if model.Exist(tools.H{"flag": r.Flag}) {
		return nil, tools.NewValidatorError(fmt.Errorf("对应平台的动态字段关系已存在，请勿重复添加"))
	}

	attr, err := tools.MapToJson(r.Attributes)
	if err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("将map转成json失败: %s", err.Error()))
	}

	frObj := model.FieldRelation{
		Flag:       r.Flag,
		Attributes: datatypes.JSON(attr),
	}

	// 创建接口
	err = model.Add(&frObj)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("创建动态字段关系失败: %s", err.Error()))
	}

	return nil, nil
}

// List 数据列表
func List(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	_, ok := req.(*model.FieldRelationListReq)
	if !ok {
		return nil, tools.ReqAssertErr
	}
	_ = c

	// 获取数据列表
	frs, err := model.List()
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("字段动态关系: %s", err.Error()))
	}

	return frs, nil
}

// Update 更新数据
func Update(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*model.FieldRelationUpdateReq)
	if !ok {
		return nil, tools.ReqAssertErr
	}
	_ = c

	filter := tools.H{"flag": r.Flag}

	if !model.Exist(filter) {
		return nil, tools.NewValidatorError(fmt.Errorf("对应平台的动态字段关系不存在"))
	}

	oldData := new(model.FieldRelation)
	err := model.Find(filter, oldData)
	if err != nil {
		return nil, tools.NewMySqlError(err)
	}

	attr, err := tools.MapToJson(r.Attributes)
	if err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("将map转成json失败: %s", err.Error()))
	}

	frObj := model.FieldRelation{
		Model:      oldData.Model,
		Flag:       r.Flag,
		Attributes: datatypes.JSON(attr),
	}

	err = model.Update(&frObj)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("更新动态字段关系失败: %s", err.Error()))
	}
	return nil, nil
}

// Delete 删除数据
func Delete(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*model.FieldRelationDeleteReq)
	if !ok {
		return nil, tools.ReqAssertErr
	}
	_ = c

	for _, id := range r.FieldRelationIds {
		filter := tools.H{"id": int(id)}
		if !model.Exist(filter) {
			return nil, tools.NewMySqlError(fmt.Errorf("动态字段关系不存在"))
		}
	}
	// 删除
	err := model.Delete(r.FieldRelationIds)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("删除动态字段关系失败: %s", err.Error()))
	}
	return nil, nil
}
