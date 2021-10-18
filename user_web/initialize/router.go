package initialize

import (
	"mytest/user_web/middlewares"
	router2 "mytest/user_web/router"

	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	router := gin.Default()

	//配置跨域
	router.Use(middlewares.Cors())
	ApiGroup := router.Group("v1")
	router2.InitUserRouter(ApiGroup)
	router2.InitBaseRouter(ApiGroup)

	return router
}
