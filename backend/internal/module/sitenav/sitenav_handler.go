package sitenav

import (
	"fmt"
	"micro-net-hub/internal/module/account/auth"
	"micro-net-hub/internal/module/sitenav/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

// GetGroups returns a list of groups.
func GetNav(c *gin.Context) {
	var navGroups = model.NewNavGroups()

	err := navGroups.FindWithSites()
	if err != nil {
		helper.ErrV2(c, helper.ReloadErr(err))
		return
	}

	helper.Success(c, navGroups)
}

type NavGroupAddReq struct {
	Title string `json:"title" validate:"required"`
	Name  string `json:"name" validate:"required"`
}

func AddNavGroup(c *gin.Context) {
	var req NavGroupAddReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取当前用户
	ctxUser, err := auth.GetCtxLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败")))
		return
	}

	g := model.NavGroup{
		Name:    req.Name,
		Title:   req.Title,
		Creator: ctxUser.Username,
	}
	err = g.Add()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("创建 NavGroup 失败: %s", err.Error())))
		return
	}

	helper.Success(c, nil)
}

type NavGroupUpdateReq struct {
	Title string `json:"title" validate:"required"`
	Name  string `json:"name" validate:"required"`
}

func UpdateNavGroup(c *gin.Context) {
	var req NavGroupUpdateReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取当前用户
	ctxUser, err := auth.GetCtxLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败")))
		return
	}

	g := &model.NavGroup{}
	err = g.FindByName(req.Name)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取 NavGroup 失败: %s", err.Error())))
		return
	}

	g.Title = req.Title
	g.Creator = ctxUser.Username
	err = g.Update()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("更新 NavGroup 失败: %s", err.Error())))
		return
	}

	helper.Success(c, nil)
}

type NavGroupDeleteReq struct {
	Ids []uint `json:"ids" validate:"required"`
}

func DeleteNavGroup(c *gin.Context) {
	var req NavGroupDeleteReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}
	g := &model.NavGroup{}
	for _, id := range req.Ids {
		err := g.FindById(id)
		if err != nil {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("查询 NavGroup 失败: %s", err.Error())))
			return
		}
		err = g.DeleteWithSites()
		if err != nil {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("删除 NavGroup 失败: %s", err.Error())))
			return
		}
	}

	helper.Success(c, nil)
}

type NavSiteAddReq struct {
	Name        string `json:"name" validate:"required"`
	NavGroupID  uint   `json:"groupid" validate:"required"`
	IconUrl     string `json:"icon"`
	Description string `json:"desc"`
	Link        string `json:"link" validate:"required"`
	DocUrl      string `json:"doc,omitempty"`
}

func AddNavSite(c *gin.Context) {
	var req NavSiteAddReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取当前用户
	ctxUser, err := auth.GetCtxLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败")))
		return
	}

	s := model.NavSite{
		Name:        req.Name,
		NavGroupID:  req.NavGroupID,
		Description: req.Description,
		Link:        req.Link,
		DocUrl:      req.DocUrl,
		IconUrl:     req.IconUrl,
		Creator:     ctxUser.Username,
	}

	err = s.Add()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("创建 NavSite 失败: %s", err.Error())))
		return
	}

	helper.Success(c, nil)
}

type NavSiteUpdateReq struct {
	ID          uint   `json:"ID" validate:"required"`
	Name        string `json:"name" validate:"required"`
	NavGroupID  uint   `json:"groupid" validate:"required"`
	IconUrl     string `json:"icon"`
	Description string `json:"desc"`
	Link        string `json:"link" validate:"required"`
	DocUrl      string `json:"doc,omitempty"`
}

func UpdateNavSite(c *gin.Context) {
	var req NavSiteUpdateReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取当前用户
	ctxUser, err := auth.GetCtxLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败")))
		return
	}

	s := &model.NavSite{}
	err = s.FindById(req.ID)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取 NavSite 失败: %s", err.Error())))
		return
	}

	s.NavGroupID = req.NavGroupID
	s.Name = req.Name
	s.Description = req.Description
	s.Link = req.Link
	s.DocUrl = req.DocUrl
	s.IconUrl = req.IconUrl
	s.Creator = ctxUser.Username

	err = s.Update()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("更新 NavSite 失败: %s", err.Error())))
		return
	}

	helper.Success(c, nil)
}

type NavSiteDeleteReq struct {
	Ids []uint `json:"ids" validate:"required"`
}

func DeleteNavSite(c *gin.Context) {
	var req NavSiteDeleteReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	for _, id := range req.Ids {
		s := &model.NavSite{}
		err := s.FindById(id)
		if err != nil {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取 NavSite 失败: %s", err.Error())))
			return
		}
		err = s.Delete()
		if err != nil {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("删除 NavSite 失败: %s", err.Error())))
			return
		}
	}
	helper.Success(c, nil)
}
