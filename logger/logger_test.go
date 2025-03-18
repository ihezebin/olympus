package logger

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
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

func BenchmarkLogger(b *testing.B) {
	ctx := context.Background()

	noOutputOpts := []Option{
		WithServiceName("unit_test"),
		WithOutput(io.Discard),
	}

	// 测试普通信息日志
	b.Run("info/plain", func(b *testing.B) {
		b.Run("logrus", func(b1 *testing.B) {
			logrusLogger := New(append(noOutputOpts, WithLoggerType(LoggerTypeLogrus))...)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				logrusLogger.Info(ctx, "hello world")
			}
		})

		b.Run("zerolog", func(b *testing.B) {
			zerologLogger := New(append(noOutputOpts, WithLoggerType(LoggerTypeZerolog))...)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				zerologLogger.Info(ctx, "hello world")
			}
		})

		b.Run("slog", func(b *testing.B) {
			slogLogger := New(append(noOutputOpts, WithLoggerType(LoggerTypeSlog))...)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				slogLogger.Info(ctx, "hello world")
			}
		})

		b.Run("zap", func(b *testing.B) {
			zapLogger := New(append(noOutputOpts, WithLoggerType(LoggerTypeZap))...)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				zapLogger.Info(ctx, "hello world")
			}
		})
	})

	// 测试带结构化字段的日志
	b.Run("info/fields", func(b *testing.B) {
		testFields := map[string]interface{}{
			"string": "value",
			"int":    123,
			"float":  3.14,
			"bool":   true,
		}

		b.Run("logrus", func(b *testing.B) {
			logrusLogger := New(append(noOutputOpts, WithLoggerType(LoggerTypeLogrus))...)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				logrusLogger.WithFields(testFields).Info(ctx, "hello world with fields")
			}
		})

		b.Run("zerolog", func(b *testing.B) {
			zerologLogger := New(append(noOutputOpts, WithLoggerType(LoggerTypeZerolog))...)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				zerologLogger.WithFields(testFields).Info(ctx, "hello world with fields")
			}
		})

		b.Run("slog", func(b *testing.B) {
			slogLogger := New(append(noOutputOpts, WithLoggerType(LoggerTypeSlog))...)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				slogLogger.WithFields(testFields).Info(ctx, "hello world with fields")
			}
		})

		b.Run("zap", func(b *testing.B) {
			zapLogger := New(append(noOutputOpts, WithLoggerType(LoggerTypeZap))...)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				zapLogger.WithFields(testFields).Info(ctx, "hello world with fields")
			}
		})
	})

}

func TestLoggerWithLocalFs(t *testing.T) {
	ctx := context.Background()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	logrusPath := filepath.Join(dir, "log/logrus.log")
	// zerologPath := filepath.Join(dir, "log/zerolog.log")
	// slogPath := filepath.Join(dir, "log/slog.log")
	// zapPath := filepath.Join(dir, "log/zap.log")

	logrusLogger := New(WithLoggerType(LoggerTypeLogrus), WithServiceName("unit_test"), WithLocalFs(LocalFsConfig{
		Path: logrusPath,
	}))
	logrusLogger.Info(ctx, "hello")
	logrusLogger.WithError(errors.New("test error")).Error(ctx, "hello")
}

func TestLoggerWithRotate(t *testing.T) {
	ctx := context.Background()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	logrusPath := filepath.Join(dir, "log_rotate/logrus.log")
	logrusLogger := New(WithLoggerType(LoggerTypeLogrus), WithServiceName("unit_test"), WithRotate(RotateConfig{
		Path:               logrusPath,
		MaxSizeKB:          10,
		MaxRetainFileCount: 3,
		MaxAge:             60 * time.Second,
		Compress:           true,
		ErrorFileLevel:     LevelError,
		ErrorFileExt:       ".err",
	}))
	for i := 0; i < 3000; i++ {
		logrusLogger.Info(ctx, "hello")
		logrusLogger.WithError(errors.New("test error")).Error(ctx, "hello")
	}
}
