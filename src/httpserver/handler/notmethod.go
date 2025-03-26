package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func HandlerMethodNotAllowed(c *gin.Context) {
	c.AbortWithStatus(http.StatusMethodNotAllowed)
}
