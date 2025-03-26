package engine

import (
	handler2 "github.com/SongZihuan/anonymous-message/src/httpserver/handler"
	"github.com/gin-gonic/gin"
)

var Engine *gin.Engine = nil

func InitEngine() error {
	gin.SetMode(gin.ReleaseMode)

	Engine = gin.New()
	Engine.Use(gin.Logger(), gin.Recovery())

	Engine.POST("/", handler2.HandlerMessage)
	Engine.POST("/message", handler2.HandlerMessage)
	Engine.POST("/message/", handler2.HandlerMessage)
	Engine.GET("/hello", handler2.HandlerHelloWorld)
	Engine.GET("/hello/", handler2.HandlerHelloWorld)

	Engine.OPTIONS("/", handler2.HandlerOptions)
	Engine.OPTIONS("/message", handler2.HandlerOptions)
	Engine.OPTIONS("/hello", handler2.HandlerOptions)

	Engine.NoRoute(handler2.HandlerMethodNotFound)
	Engine.NoMethod(handler2.HandlerMethodNotAllowed)

	return nil
}
