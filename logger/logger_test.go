package logger

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

func TestLogger(t *testing.T) {
	ctx := context.Background()

	logrusLogger := New(WithLoggerType(LoggerTypeLogrus), WithServiceName("unit_test"), WithLevel(LevelDebug))
	logrusLogger.Debug(ctx, "hello")

	zerologLogger := New(WithLoggerType(LoggerTypeZerolog), WithServiceName("unit_test"), WithLevel(LevelDebug))
	zerologLogger.Debug(ctx, "hello")

	slogLogger := New(WithLoggerType(LoggerTypeSlog), WithServiceName("unit_test"), WithLevel(LevelDebug))
	slogLogger.Debug(ctx, "hello")

	zapLogger := New(WithLoggerType(LoggerTypeZap), WithServiceName("unit_test"), WithLevel(LevelDebug))
	zapLogger.Debug(ctx, "hello")
}

func TestLoggerWithTraceId(t *testing.T) {
	ctx := context.Background()
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		t.Errorf("OTEL ERROR: %v", err)
	}))
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	exporter, err := otlploghttp.New(ctx,
		otlploghttp.WithInsecure(),
		otlploghttp.WithEndpoint("localhost:4318"),
	)
	if err != nil {
		t.Fatalf("无法创建 OTLP log HTTP exporter: %v", err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName("test_logger_with_trace_id"),
	)
	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)
	global.SetLoggerProvider(provider)

	// 获取 Tracer
	tracer := otel.Tracer("github.com/ihezebin/olympus/logger")
	ctx, span := tracer.Start(ctx, "unit_test")
	defer span.End()

	logrusLogger := New(WithLoggerType(LoggerTypeLogrus), WithServiceName("unit_test"), WithOtlpEnabled(true))
	logrusLogger.Info(ctx, "hello")

	zerologLogger := New(WithLoggerType(LoggerTypeZerolog), WithServiceName("unit_test"), WithOtlpEnabled(true))
	zerologLogger.Info(ctx, "hello")

	slogLogger := New(WithLoggerType(LoggerTypeSlog), WithServiceName("unit_test"), WithOtlpEnabled(true))
	slogLogger.Info(ctx, "hello")

	zapLogger := New(WithLoggerType(LoggerTypeZap), WithServiceName("unit_test"), WithOtlpEnabled(true))
	zapLogger.Info(ctx, "hello")

	time.Sleep(10 * time.Second)
}

func TestLoggerError(t *testing.T) {
	ctx := context.Background()

	err := errors.New("test error")

	logrusLogger := New(WithLoggerType(LoggerTypeLogrus), WithServiceName("unit_test"))
	logrusLogger.WithError(err).Error(ctx, "hello err")

	zerologLogger := New(WithLoggerType(LoggerTypeZerolog), WithServiceName("unit_test"))
	zerologLogger.WithError(err).Error(ctx, "hello err")

	slogLogger := New(WithLoggerType(LoggerTypeSlog), WithServiceName("unit_test"))
	slogLogger.WithError(err).Error(ctx, "hello err")

	zapLogger := New(WithLoggerType(LoggerTypeZap), WithServiceName("unit_test"))
	zapLogger.WithError(err).Error(ctx, "hello err")
}

/*
BenchmarkLogger/info/default/logrus-8     393117              3070 ns/op            1762 B/op         34 allocs/op
BenchmarkLogger/info/default/zerolog-8    715218              1744 ns/op            1977 B/op         15 allocs/op
BenchmarkLogger/info/default/slog-8       602392              1900 ns/op             432 B/op         11 allocs/op
BenchmarkLogger/info/default/zap-8        844246              1512 ns/op             825 B/op          9 allocs/op

zap > slog > zerolog > logrus
*/
func BenchmarkLogger(b *testing.B) {
	ctx := context.Background()

	noOutputOpts := []Option{
		WithServiceName("unit_test"),
		WithOutput(io.Discard),
	}

	// 测试普通信息日志
	b.Run("info/default", func(b *testing.B) {
		b.Run("logrus", func(b *testing.B) {
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

	// 测试是否开启 otlp 的性能差异
	b.Run("otlp/enabled", func(b *testing.B) {
		otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		}))
		tp := trace.NewTracerProvider()
		otel.SetTracerProvider(tp)
		exporter, err := otlploghttp.New(ctx,
			otlploghttp.WithInsecure(),
			otlploghttp.WithEndpoint("localhost:4318"),
		)
		if err != nil {
			b.Fatalf("无法创建 OTLP log HTTP exporter: %v", err)
		}

		res := resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("test_logger_with_trace_id"),
		)
		provider := sdklog.NewLoggerProvider(
			sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
			sdklog.WithResource(res),
		)
		global.SetLoggerProvider(provider)

		// 获取 Tracer
		tracer := otel.Tracer("github.com/ihezebin/olympus/logger")
		ctx, span := tracer.Start(ctx, "unit_test")
		defer span.End()
		b.Run("default", func(b *testing.B) {
			logrusLogger := New(append(noOutputOpts, WithLoggerType(LoggerTypeZap), WithServiceName("unit_test"))...)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				logrusLogger.Info(ctx, "hello")
			}
		})

		b.Run("otlp", func(b *testing.B) {
			logrusLogger := New(append(noOutputOpts, WithLoggerType(LoggerTypeZap), WithServiceName("unit_test"), WithOtlpEnabled(true))...)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				logrusLogger.Info(ctx, "hello")
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
	zerologPath := filepath.Join(dir, "log/zerolog.log")
	slogPath := filepath.Join(dir, "log/slog.log")
	zapPath := filepath.Join(dir, "log/zap.log")

	logrusLogger := New(WithLoggerType(LoggerTypeLogrus), WithServiceName("unit_test"), WithLocalFs(LocalFsConfig{
		Path: logrusPath,
	}))
	logrusLogger.WithField("a", "1").WithField("b", "2").Info(ctx, "hello")
	logrusLogger.WithError(errors.New("test error")).Error(ctx, "hello")

	zerologLogger := New(WithLoggerType(LoggerTypeZerolog), WithServiceName("unit_test"), WithLocalFs(LocalFsConfig{
		Path: zerologPath,
	}))
	zerologLogger.WithField("a", "1").WithField("b", "2").Info(ctx, "hello")
	zerologLogger.WithError(errors.New("test error")).Error(ctx, "hello")

	slogLogger := New(WithLoggerType(LoggerTypeSlog), WithServiceName("unit_test"), WithLocalFs(LocalFsConfig{
		Path: slogPath,
	}))
	slogLogger.WithField("a", "1").WithField("b", "2").Info(ctx, "hello")
	slogLogger.WithError(errors.New("test error")).Error(ctx, "hello")

	zapLogger := New(WithLoggerType(LoggerTypeZap), WithServiceName("unit_test"), WithLocalFs(LocalFsConfig{
		Path: zapPath,
	}))
	zapLogger.WithField("a", "1").WithField("b", "2").Info(ctx, "hello")
	zapLogger.WithError(errors.New("test error")).Error(ctx, "hello")
}

func TestLogrusLoggerWithRotate(t *testing.T) {
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
		logrusLogger.WithField("a", "1").WithField("b", "2").Info(ctx, "hello")
		logrusLogger.WithError(errors.New("test error")).Error(ctx, "hello err")
	}
}

func TestZerologLoggerWithRotate(t *testing.T) {
	ctx := context.Background()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	zerologPath := filepath.Join(dir, "log_rotate/zerolog.log")
	zerologLogger := New(WithLoggerType(LoggerTypeZerolog), WithServiceName("unit_test"), WithRotate(RotateConfig{
		Path:               zerologPath,
		MaxSizeKB:          10,
		MaxRetainFileCount: 3,
		MaxAge:             60 * time.Second,
		Compress:           true,
		ErrorFileLevel:     LevelError,
		ErrorFileExt:       ".err",
	}))

	for i := 0; i < 3000; i++ {
		zerologLogger.WithField("a", "1").WithField("b", "2").Info(ctx, "hello")
		zerologLogger.WithError(errors.New("test error")).Error(ctx, "hello err")
	}
}

func TestSlogLoggerWithRotate(t *testing.T) {
	ctx := context.Background()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	slogPath := filepath.Join(dir, "log_rotate/slog.log")
	slogLogger := New(WithLoggerType(LoggerTypeSlog), WithServiceName("unit_test"), WithRotate(RotateConfig{
		Path:               slogPath,
		MaxSizeKB:          10,
		MaxRetainFileCount: 3,
		MaxAge:             60 * time.Second,
		Compress:           true,
		ErrorFileLevel:     LevelError,
		ErrorFileExt:       ".err",
	}))

	for i := 0; i < 3000; i++ {
		slogLogger.WithField("a", "1").Info(ctx, "hello")
		slogLogger.WithError(errors.New("test error")).Error(ctx, "hello err")
	}
}

func TestZapLoggerWithRotate(t *testing.T) {
	ctx := context.Background()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	zapPath := filepath.Join(dir, "log_rotate/zap.log")
	zapLogger := New(WithLoggerType(LoggerTypeZap), WithServiceName("unit_test"), WithRotate(RotateConfig{
		Path:               zapPath,
		MaxSizeKB:          10,
		MaxRetainFileCount: 3,
		MaxAge:             60 * time.Second,
		Compress:           true,
		ErrorFileLevel:     LevelError,
		ErrorFileExt:       ".err",
	}))

	for i := 0; i < 3000; i++ {
		zapLogger.WithField("a", "1").WithField("b", "2").Info(ctx, "hello")
		zapLogger.WithError(errors.New("test error")).Error(ctx, "hello err")
	}
}
