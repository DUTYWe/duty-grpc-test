package router

import (
	"mytest/user_web/api"

	"github.com/gin-gonic/gin"
)

func InitBaseRouter(router *gin.RouterGroup) {
	BaseRouter := router.Group("base")
	{
		BaseRouter.GET("captcha", api.Getcaptcha)
	}
}
