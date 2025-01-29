package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func HandlerMethodNotFound(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotFound)
}
