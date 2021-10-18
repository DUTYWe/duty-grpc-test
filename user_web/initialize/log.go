package initialize

import "go.uber.org/zap"

func InitLogger() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	// logger,_ := zap.NewProduction()
	// defer logger.Sync()
	// suger := logger.Sugar()
	/*
		1、S可以获取一个全局的suger，可以让我们设计一个全局的logger
		2、日志是分级别的，debug,info,warn,error,fetal,从左到右由低到高，当配置到info级别，则打印不出来debug级别。
		3、S和L函数常用，给我们提供了一个全局的安全访问logger的途径
	*/
}
