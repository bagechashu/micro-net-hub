package sitenav

import (
	"fmt"
	"micro-net-hub/internal/module/sitenav/model"
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

func ListNav(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*model.NavReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c

	var err error
	var gs = model.NewNavGroups()
	err = gs.Find(r)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取接口列表失败: %s", err.Error()))
	}

	var sites = model.NewNavSites()
	err = sites.Find(r)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取接口列表失败: %s", err.Error()))
	}

	gCount, err := model.GroupCount()
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取接口列表失败: %s", err.Error()))
	}
	sCount, err := model.SiteCount()
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取接口列表失败: %s", err.Error()))
	}

	return model.NavRsp{
		Groups:     *gs,
		GroupTotal: gCount,
		Sites:      *sites,
		SiteTotal:  sCount,
	}, nil
}

func AddNavGroup(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	return nil, nil
}

func UpdateNavGroup(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	return nil, nil
}

func DeleteNavGroup(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	return nil, nil
}

func AddNavSite(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	return nil, nil
}

func UpdateNavSite(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	return nil, nil
}

func DeleteNavSite(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	return nil, nil
}
