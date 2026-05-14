// Package middleware 见文知意
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func T(c *gin.Context) {
	c.JSON(http.StatusOK, "t")
}
