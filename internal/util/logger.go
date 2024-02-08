package util

import "go.uber.org/zap"

func GetLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	defer logger.Sync()
	if err != nil {
		panic(err)
	}
	return logger
}
