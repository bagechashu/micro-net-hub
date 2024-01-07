package apimgr

import (
	"fmt"

	apiMgrModel "micro-net-hub/internal/module/apimgr/model"
	userLogic "micro-net-hub/internal/module/user"
	"micro-net-hub/internal/server/helper"
	"micro-net-hub/internal/tools"

	"github.com/gin-gonic/gin"
	"github.com/thoas/go-funk"
)

// Add 添加数据
func Add(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*apiMgrModel.ApiAddReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c

	// 获取当前用户
	ctxUser, err := userLogic.GetCurrentLoginUser(c)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败"))
	}

	api := apiMgrModel.Api{
		Method:   r.Method,
		Path:     r.Path,
		Category: r.Category,
		Remark:   r.Remark,
		Creator:  ctxUser.Username,
	}

	// 创建接口
	err = apiMgrModel.Add(&api)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("创建接口失败: %s", err.Error()))
	}

	return nil, nil
}

// List 数据列表
func List(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*apiMgrModel.ApiListReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c

	// 获取数据列表
	apis, err := apiMgrModel.List(r)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取接口列表失败: %s", err.Error()))
	}

	rets := make([]apiMgrModel.Api, 0)
	for _, api := range apis {
		rets = append(rets, *api)
	}
	count, err := apiMgrModel.Count()
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取接口总数失败"))
	}

	return apiMgrModel.ApiListRsp{
		Total: count,
		Apis:  rets,
	}, nil
}

// GetTree 数据树
func GetTree(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*apiMgrModel.ApiGetTreeReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c
	_ = r

	apis, err := apiMgrModel.ListAll()
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取资源列表失败: " + err.Error()))
	}

	// 获取所有的分类
	var categoryList []string
	for _, api := range apis {
		categoryList = append(categoryList, api.Category)
	}
	// 获取去重后的分类
	categoryUniq := funk.UniqString(categoryList)

	apiTree := make([]*apiMgrModel.ApiTreeRsp, len(categoryUniq))

	for i, category := range categoryUniq {
		apiTree[i] = &apiMgrModel.ApiTreeRsp{
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

	return apiTree, nil
}

// Update 更新数据
func Update(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*apiMgrModel.ApiUpdateReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c

	filter := tools.H{"id": int(r.ID)}
	if !apiMgrModel.Exist(filter) {
		return nil, helper.NewMySqlError(fmt.Errorf("接口不存在"))
	}

	// 获取当前登陆用户
	ctxUser, err := userLogic.GetCurrentLoginUser(c)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户失败"))
	}

	oldData := new(apiMgrModel.Api)
	err = apiMgrModel.Find(filter, oldData)
	if err != nil {
		return nil, helper.NewMySqlError(err)
	}

	api := apiMgrModel.Api{
		Model:    oldData.Model,
		Method:   r.Method,
		Path:     r.Path,
		Category: r.Category,
		Remark:   r.Remark,
		Creator:  ctxUser.Username,
	}
	err = apiMgrModel.Update(&api)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("更新接口失败: %s", err.Error()))
	}
	return nil, nil
}

// Delete 删除数据
func Delete(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*apiMgrModel.ApiDeleteReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c

	for _, id := range r.ApiIds {
		filter := tools.H{"id": int(id)}
		if !apiMgrModel.Exist(filter) {
			return nil, helper.NewMySqlError(fmt.Errorf("接口不存在"))
		}
	}
	// 删除接口
	err := apiMgrModel.Delete(r.ApiIds)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("删除接口失败: %s", err.Error()))
	}
	return nil, nil
}
