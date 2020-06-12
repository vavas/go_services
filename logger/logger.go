package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a global logger object.
var Logger *zap.Logger

// Init global logger
func InitLogging(serviceName string) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := config.Build()

	// Add service name
	Logger = logger.Named("service:" + serviceName)
}
