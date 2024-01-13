package handler

import (
	siteNavLogic "micro-net-hub/internal/module/sitenav"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

type SiteNavHandler struct{}

// List 记录列表
func (SiteNavHandler) GetAllNavConfig(c *gin.Context) {
	helper.BindAndValidateRequest(c, nil)

	data, respErr := siteNavLogic.GetAllNavConfig(c, nil)
	helper.HandleResponse(c, data, respErr)
}

// List 记录列表
func (SiteNavHandler) GetSideNavGroups(c *gin.Context) {
	helper.BindAndValidateRequest(c, nil)

	data, respErr := siteNavLogic.GetSideNavGroups(c, nil)
	helper.HandleResponse(c, data, respErr)
}

// List 记录列表
func (SiteNavHandler) GetNavGroups(c *gin.Context) {
	helper.BindAndValidateRequest(c, nil)

	data, respErr := siteNavLogic.GetNavGroups(c, nil)
	helper.HandleResponse(c, data, respErr)
}
