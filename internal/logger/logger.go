package logger

import (
	"fmt"
	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitLogger(level string) {
	var config zap.Config

	switch level {
	case "debug":
		config = zap.NewDevelopmentConfig()
	case "info", "warn", "error":
		config = zap.NewProductionConfig()
		config.Level = getZapLevel(level)
	default:
		config = zap.NewProductionConfig()
	}

	var err error
	Logger, err = config.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
}

func getZapLevel(level string) zap.AtomicLevel {
	switch level {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
}

func GetLogger() *zap.Logger {
	return Logger
}
