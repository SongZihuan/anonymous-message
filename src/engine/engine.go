package engine

import (
	"github.com/SongZihuan/anonymous-message/src/handler"
	"github.com/gin-gonic/gin"
)

var Engine *gin.Engine = nil

func InitEngine() error {
	gin.SetMode(gin.ReleaseMode)

	Engine = gin.New()
	Engine.Use(gin.Logger(), gin.Recovery())

	Engine.POST("/", handler.HandlerMessage)
	Engine.POST("/message", handler.HandlerMessage)
	Engine.POST("/message/", handler.HandlerMessage)
	Engine.GET("/hello", handler.HandlerHelloWorld)
	Engine.GET("/hello/", handler.HandlerHelloWorld)

	Engine.OPTIONS("/", handler.HandlerOptions)
	Engine.OPTIONS("/message", handler.HandlerOptions)
	Engine.OPTIONS("/hello", handler.HandlerOptions)

	Engine.NoRoute(handler.HandlerMethodNotFound)
	Engine.NoMethod(handler.HandlerMethodNotAllowed)

	return nil
}
