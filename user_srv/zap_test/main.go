package main

import (
	"go.uber.org/zap"
)

func main() {
	//logger, _ := zap.NewProduction()  //输出json文件
	logger, _ := zap.NewDevelopment() //输出日志
	defer logger.Sync()               // flushes buffer, if any刷新缓存
	//url := "https://imooc.com"
	logger.Info("failed to fetch URL")
	// sugar := logger.Sugar()
	// sugar.Infow("failed to fetch URL",
	// 	// Structured context as loosely typed key-value pairs.
	// 	"url", url,
	// 	"attempt", 3,
	// 	"backoff", time.Second,
	// )
	// sugar.Infof("Failed to fetch URL: %s", url)
}
