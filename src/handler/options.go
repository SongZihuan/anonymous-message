package handler

import (
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"github.com/SongZihuan/anonymous-message/src/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func HandlerOptions(c *gin.Context) {
	ok := handlerOptions(c)
	if ok {
		c.Writer.WriteHeader(http.StatusNoContent)
		return
	} else {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
}

func handlerOptions(c *gin.Context) bool {
	if flagparser.Origin != "" {
		origin, ok := checkOrigin(c)
		if ok {
			allowHeaderWriter(c, origin)
			return true
		} else {
			c.Writer.Header().Del("Access-Control-Allow-Origin") // 确保没有此请求头
			return false
		}
	} else {
		allowHeaderWriter(c, utils.OriginClear(c.GetHeader("Origin")))
		return true
	}
}

func allowHeaderWriter(c *gin.Context, origin string) {
	if origin == "" {
		origin = "*"
	}

	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "1728000") // 此处单位秒，20天
}

func checkOrigin(c *gin.Context) (string, bool) {
	origin := utils.OriginClear(c.GetHeader("Origin"))
	if origin == "" {
		return "", false
	}

	for _, o := range strings.Split(origin, ",") {
		if o == "*" {
			return origin, true
		} else if o = utils.OriginClear(o); o == origin { // origin肯定不会是""，因此此处不用判断o是否为""
			return origin, true
		}
	}
	return "", false
}
