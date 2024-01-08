package handler

import (
	"micro-net-hub/internal/tools"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Demo(c *gin.Context) {
	c.JSON(http.StatusOK, tools.H{"code": 200, "msg": "ok", "data": "pong"})
}
