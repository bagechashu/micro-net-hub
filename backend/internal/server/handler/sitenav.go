package handler

import (
	siteNavLogic "micro-net-hub/internal/module/sitenav"
	siteNavModel "micro-net-hub/internal/module/sitenav/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

type SiteNavHandler struct{}

func (SiteNavHandler) GetNav(c *gin.Context) {
	req := new(helper.EmptyStruct)
	helper.HandleRequest(c, req, siteNavLogic.GetNav)
}

func (SiteNavHandler) AddNavGroup(c *gin.Context) {
	req := new(siteNavModel.NavGroupAddReq)
	helper.HandleRequest(c, req, siteNavLogic.AddNavGroup)
}

func (SiteNavHandler) UpdateNavGroup(c *gin.Context) {
	req := new(siteNavModel.NavGroupUpdateReq)
	helper.HandleRequest(c, req, siteNavLogic.UpdateNavGroup)
}

func (SiteNavHandler) DeleteNavGroup(c *gin.Context) {
	req := new(siteNavModel.NavGroupDeleteReq)
	helper.HandleRequest(c, req, siteNavLogic.DeleteNavGroup)
}

func (SiteNavHandler) AddNavSite(c *gin.Context) {
	req := new(siteNavModel.NavSiteAddReq)
	helper.HandleRequest(c, req, siteNavLogic.AddNavSite)
}

func (SiteNavHandler) UpdateNavSite(c *gin.Context) {
	req := new(siteNavModel.NavSiteUpdateReq)
	helper.HandleRequest(c, req, siteNavLogic.UpdateNavSite)
}

func (SiteNavHandler) DeleteNavSite(c *gin.Context) {
	req := new(siteNavModel.NavSiteDeleteReq)
	helper.HandleRequest(c, req, siteNavLogic.DeleteNavSite)
}
