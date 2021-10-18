package main

import (
	"go.uber.org/zap"
)

func main() {
	logger, err := NewLogger()
	if err != nil {
		panic(err)
	}
	suger := logger.Sugar()
	defer suger.Sync()
	url := "https://imooc.com"
	logger.Info("failed to fetch URL",
		zap.String("url", url),
		zap.Int("attempt", 3),
	)
	// sugar := logger.Sugar()
	// sugar.Infow("failed to fetch URL",
	// 	// Structured context as loosely typed key-value pairs.
	// 	"url", url,
	// 	"attempt", 3,
	// 	"backoff", time.Second,
	// )
	// sugar.Infof("Failed to fetch URL: %s", url)
}

func NewLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{
		"./mytest.log",
		"stderr",
		"stdout",
	}
	return cfg.Build()
}
