package sitenav

import (
	"fmt"
	"micro-net-hub/internal/module/sitenav/model"
	userLogic "micro-net-hub/internal/module/user"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

// GetGroups returns a list of groups.
func GetNav(c *gin.Context) {
	var navGroups = model.NewNavGroups()

	err := navGroups.FindWithSites()
	if err != nil {
		helper.Err(c, helper.ReloadErr(err), nil)
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
	ctxUser, err := userLogic.GetCurrentLoginUser(c)
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
	req := new(NavGroupUpdateReq)
	helper.HandleRequest(
		c,
		req,
		func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
			r, ok := req.(*NavGroupUpdateReq)
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
		})
}

type NavGroupDeleteReq struct {
	Ids []uint `json:"ids" validate:"required"`
}

func DeleteNavGroup(c *gin.Context) {
	req := new(NavGroupDeleteReq)
	helper.HandleRequest(
		c,
		req,
		func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
			r, ok := req.(*NavGroupDeleteReq)
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
		})
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
	req := new(NavSiteAddReq)
	helper.HandleRequest(
		c,
		req,
		func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
			r, ok := req.(*NavSiteAddReq)
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
		})
}

type NavSiteUpdateReq struct {
	Name        string `json:"name" validate:"required"`
	NavGroupID  uint   `json:"groupid" validate:"required"`
	IconUrl     string `json:"icon"`
	Description string `json:"desc"`
	Link        string `json:"link" validate:"required"`
	DocUrl      string `json:"doc,omitempty"`
}

func UpdateNavSite(c *gin.Context) {
	req := new(NavSiteUpdateReq)
	helper.HandleRequest(
		c,
		req,
		func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
			r, ok := req.(*NavSiteUpdateReq)
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
		})
}

type NavSiteDeleteReq struct {
	Ids []uint `json:"ids" validate:"required"`
}

func DeleteNavSite(c *gin.Context) {
	req := new(NavSiteDeleteReq)
	helper.HandleRequest(
		c,
		req,
		func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
			r, ok := req.(*NavSiteDeleteReq)
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
		})
}
