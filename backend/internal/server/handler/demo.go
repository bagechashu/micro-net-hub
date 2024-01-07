package handler

import (
	"micro-net-hub/internal/tools"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	Api           = &ApiController{}
	Group         = &GroupController{}
	Menu          = &MenuController{}
	Role          = &RoleController{}
	User          = &UserController{}
	OperationLog  = &OperationLogController{}
	Base          = &BaseController{}
	FieldRelation = &FieldRelationController{}
)

func Demo(c *gin.Context) {
	c.JSON(http.StatusOK, tools.H{"code": 200, "msg": "ok", "data": "pong"})
}
