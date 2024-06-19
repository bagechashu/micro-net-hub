package demo

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Demo(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{"code": 200, "msg": "ok", "data": "pong"})
}
