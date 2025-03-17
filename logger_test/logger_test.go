package logger_test

import (
	"context"
	"testing"

	"github.com/ihezebin/olympus/logger"
)

func TestLogger(t *testing.T) {
	ctx := context.Background()

	logrusLogger := logger.New(logger.WithLoggerType(logger.LoggerTypeLogrus), logger.WithServiceName("unit_test"))
	logrusLogger.Info(ctx, "hello")

	zerologLogger := logger.New(logger.WithLoggerType(logger.LoggerTypeZerolog), logger.WithServiceName("unit_test"))
	zerologLogger.Info(ctx, "hello")

	slogLogger := logger.New(logger.WithLoggerType(logger.LoggerTypeSlog), logger.WithServiceName("unit_test"))
	slogLogger.Info(ctx, "hello")

	zapLogger := logger.New(logger.WithLoggerType(logger.LoggerTypeZap), logger.WithServiceName("unit_test"))
	zapLogger.Info(ctx, "hello")
}
