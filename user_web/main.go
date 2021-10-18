package main

import (
	"fmt"
	"mytest/user_web/global"
	"mytest/user_web/initialize"
	myvalidator "mytest/user_web/validator"

	ut "github.com/go-playground/universal-translator"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

func main() {
	//1、初始化router
	router := initialize.Routers()
	//2、初始化logger
	initialize.InitLogger()
	//3、初始化配置文件
	initialize.InitConfig()
	//4、初始化翻译
	initialize.InitTrans("zh")

	//注册验证器
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("mobile", myvalidator.ValidateMobile)
		_ = v.RegisterTranslation("mobile", global.Trans, func(ut ut.Translator) error {
			return ut.Add("mobile", "{0} 非法的手机号码！", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("mobile", fe.Field())
			return t
		})
	}
	zap.S().Infof("启动服务器，端口：%d", global.ServerConfig.Port)

	if err := router.Run(fmt.Sprintf(":%d", global.ServerConfig.Port)); err != nil {
		zap.S().Panic("启动失败：", err.Error())
	}
}
