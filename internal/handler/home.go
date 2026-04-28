package handler

import (
	"net/http"
	"web/internal/types"

	"github.com/gin-gonic/gin"
)

func Home() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(types.HomeTmpl))
	}
}
