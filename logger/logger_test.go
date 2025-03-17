package logger

import (
	"context"
	"testing"
)

func TestLogger(t *testing.T) {
	ctx := context.Background()

	logrusLogger := New(WithLoggerType(LoggerTypeLogrus), WithServiceName("unit_test"))
	logrusLogger.Info(ctx, "hello")

	zerologLogger := New(WithLoggerType(LoggerTypeZerolog), WithServiceName("unit_test"))
	zerologLogger.Info(ctx, "hello")

	slogLogger := New(WithLoggerType(LoggerTypeSlog), WithServiceName("unit_test"))
	slogLogger.Info(ctx, "hello")

	zapLogger := New(WithLoggerType(LoggerTypeZap), WithServiceName("unit_test"))
	zapLogger.Info(ctx, "hello")
}
