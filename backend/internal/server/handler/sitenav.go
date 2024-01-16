package handler

import (
	siteNavLogic "micro-net-hub/internal/module/sitenav"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

type SiteNavHandler struct{}

// List 记录列表
func (SiteNavHandler) GetNavSites(c *gin.Context) {
	helper.BindAndValidateRequest(c, nil)

	data, respErr := siteNavLogic.GetNavSites(c, nil)
	helper.HandleResponse(c, data, respErr)
}
