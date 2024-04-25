package apimgr

import (
	"micro-net-hub/internal/module/account/current"
	"micro-net-hub/internal/module/apimgr/model"

	"github.com/gin-gonic/gin"

	"fmt"

	"micro-net-hub/internal/server/helper"
	"micro-net-hub/internal/tools"

	"github.com/thoas/go-funk"
)

// ApiListReq 获取资源列表结构体
type ApiListReq struct {
	Method   string `json:"method" form:"method"`
	Path     string `json:"path" form:"path"`
	Category string `json:"category" form:"category"`
	Creator  string `json:"creator" form:"creator"`
	PageNum  int    `json:"pageNum" form:"pageNum"`
	PageSize int    `json:"pageSize" form:"pageSize"`
}

type ApiListRsp struct {
	Total int64       `json:"total"`
	Apis  []model.Api `json:"apis"`
}

// List 记录列表
func List(c *gin.Context) {
	var req ApiListReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取数据列表
	apis, err := model.List(
		&model.Api{
			Method:   req.Method,
			Path:     req.Path,
			Category: req.Category,
			Creator:  req.Creator,
		},
		req.PageNum,
		req.PageSize)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取接口列表失败: %s", err.Error())))
		return
	}

	rets := make([]model.Api, 0)
	for _, api := range apis {
		rets = append(rets, *api)
	}
	count, err := model.Count()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取接口总数失败")))
		return
	}

	helper.Success(c, ApiListRsp{
		Total: count,
		Apis:  rets,
	})
}

// ApiGetTreeReq 获取资源树结构体
type ApiTreeRsp struct {
	ID       int          `json:"ID"`
	Remark   string       `json:"remark"`
	Category string       `json:"category"`
	Children []*model.Api `json:"children"`
}

// GetTree 接口树
func GetTree(c *gin.Context) {
	apis, err := model.ListAll()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取资源列表失败: "+err.Error())))
		return
	}

	// 获取所有的分类
	var categoryList []string
	for _, api := range apis {
		categoryList = append(categoryList, api.Category)
	}
	// 获取去重后的分类
	categoryUniq := funk.UniqString(categoryList)

	apiTree := make([]*ApiTreeRsp, len(categoryUniq))

	for i, category := range categoryUniq {
		apiTree[i] = &ApiTreeRsp{
			ID:       -i,
			Remark:   category,
			Category: category,
			Children: nil,
		}
		for _, api := range apis {
			if category == api.Category {
				apiTree[i].Children = append(apiTree[i].Children, api)
			}
		}
	}

	helper.Success(c, apiTree)
}

// ApiAddReq 添加资源结构体
type ApiAddReq struct {
	Method   string `json:"method" validate:"required,min=1,max=20"`
	Path     string `json:"path" validate:"required,min=1,max=100"`
	Category string `json:"category" validate:"required,min=1,max=50"`
	Remark   string `json:"remark" validate:"min=0,max=100"`
}

// Add 新建记录
func Add(c *gin.Context) {
	var req ApiAddReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取当前用户
	ctxUser, err := current.GetCurrentLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败")))
		return
	}

	api := model.Api{
		Method:   req.Method,
		Path:     req.Path,
		Category: req.Category,
		Remark:   req.Remark,
		Creator:  ctxUser.Username,
	}

	// 创建接口
	err = model.Add(&api)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("创建接口失败: %s", err.Error())))
		return
	}

	helper.Success(c, nil)
}

// ApiUpdateReq 更新资源结构体
type ApiUpdateReq struct {
	ID       uint   `json:"id" validate:"required"`
	Method   string `json:"method" validate:"min=1,max=20"`
	Path     string `json:"path" validate:"min=1,max=100"`
	Category string `json:"category" validate:"min=1,max=50"`
	Remark   string `json:"remark" validate:"min=0,max=100"`
}

// Update 更新记录
func Update(c *gin.Context) {
	var req ApiUpdateReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	filter := tools.H{"id": int(req.ID)}
	if !model.Exist(filter) {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("接口不存在")))
		return
	}

	// 获取当前登陆用户
	ctxUser, err := current.GetCurrentLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户失败")))
		return
	}

	oldData := new(model.Api)
	err = model.Find(filter, oldData)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(err))
		return
	}

	api := model.Api{
		Model:    oldData.Model,
		Method:   req.Method,
		Path:     req.Path,
		Category: req.Category,
		Remark:   req.Remark,
		Creator:  ctxUser.Username,
	}
	err = model.Update(&api)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("更新接口失败: %s", err.Error())))
		return
	}
	helper.Success(c, nil)
}

// ApiDeleteReq 删除资源结构体
type ApiDeleteReq struct {
	ApiIds []uint `json:"apiIds" validate:"required"`
}

// Delete 删除记录
func Delete(c *gin.Context) {
	var req ApiDeleteReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	for _, id := range req.ApiIds {
		filter := tools.H{"id": int(id)}
		if !model.Exist(filter) {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("接口不存在")))
			return
		}
	}
	// 删除接口
	err = model.Delete(req.ApiIds)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("删除接口失败: %s", err.Error())))
		return
	}
	helper.Success(c, nil)
}
