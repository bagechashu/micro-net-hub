package sitenav

import (
	"fmt"
	"micro-net-hub/internal/module/sitenav/model"
	userLogic "micro-net-hub/internal/module/user"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

// GetGroups returns a list of groups.
func GetNav(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	var navGroups = model.NewNavGroups()
	if err := navGroups.FindWithSites(); err != nil {
		return nil, err
	}

	return navGroups, nil
}

func AddNavGroup(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*model.NavGroupAddReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}

	// 获取当前用户
	ctxUser, err := userLogic.GetCurrentLoginUser(c)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败"))
	}

	g := model.NavGroup{
		Name:    r.Name,
		Title:   r.Title,
		Creator: ctxUser.Username,
	}

	err = g.Add()
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("创建 NavGroup 失败: %s", err.Error()))
	}

	return nil, nil
}

func UpdateNavGroup(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*model.NavGroupUpdateReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}

	g := &model.NavGroup{}
	err := g.FindByName(r.Name)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取 NavGroup 失败: %s", err.Error()))
	}
	// 获取当前用户
	ctxUser, err := userLogic.GetCurrentLoginUser(c)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败"))
	}

	g.Title = r.Title
	g.Creator = ctxUser.Username

	err = g.Update()
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("更新 NavGroup 失败: %s", err.Error()))
	}

	return nil, nil
}

func DeleteNavGroup(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*model.NavGroupDeleteReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}

	g := &model.NavGroup{}
	for _, id := range r.Ids {
		err := g.FindById(id)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("查询 NavGroup 失败: %s", err.Error()))
		}
		err = g.DeleteWithSites()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("删除 NavGroup 失败: %s", err.Error()))
		}
	}

	return nil, nil
}

func AddNavSite(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*model.NavSiteAddReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}

	// 获取当前用户
	ctxUser, err := userLogic.GetCurrentLoginUser(c)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败"))
	}

	s := model.NavSite{
		Name:        r.Name,
		NavGroupID:  r.NavGroupID,
		Description: r.Description,
		Link:        r.Link,
		DocUrl:      r.DocUrl,
		IconUrl:     r.IconUrl,
		Creator:     ctxUser.Username,
	}

	err = s.Add()
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("创建 NavSite 失败: %s", err.Error()))
	}

	return nil, nil
}

func UpdateNavSite(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*model.NavSiteUpdateReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}

	s := &model.NavSite{}
	err := s.FindByName(r.Name)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取 NavSite 失败: %s", err.Error()))
	}
	// 获取当前用户
	ctxUser, err := userLogic.GetCurrentLoginUser(c)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败"))
	}

	s.NavGroupID = r.NavGroupID
	s.Description = r.Description
	s.Link = r.Link
	s.DocUrl = r.DocUrl
	s.IconUrl = r.IconUrl
	s.Creator = ctxUser.Username

	err = s.Update()
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("更新 NavSite 失败: %s", err.Error()))
	}

	return nil, nil
}

func DeleteNavSite(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*model.NavSiteDeleteReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}

	for _, id := range r.Ids {
		s := &model.NavSite{}
		err := s.FindById(id)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取 NavSite 失败: %s", err.Error()))
		}
		err = s.Delete()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("删除 NavSite 失败: %s", err.Error()))
		}
	}
	return nil, nil
}
