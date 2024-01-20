package handler

import (
	siteNavLogic "micro-net-hub/internal/module/sitenav"
	siteNavModel "micro-net-hub/internal/module/sitenav/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

type SiteNavHandler struct{}

// List 记录列表
func (SiteNavHandler) GetNav(c *gin.Context) {
	helper.BindAndValidateRequest(c, nil)

	data, respErr := siteNavLogic.GetNav(c, nil)
	helper.HandleResponse(c, data, respErr)
}

// List for manager
func (SiteNavHandler) ListNav(c *gin.Context) {
	req := new(siteNavModel.NavReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := siteNavLogic.ListNav(c, req)
	helper.HandleResponse(c, data, respErr)
}
