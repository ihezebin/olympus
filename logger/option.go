package logger

import (
	"context"
	"io"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type Options struct {
	Type        LoggerType
	Level       Level
	ServiceName string
	Output      io.Writer
	// LocalFsConfig 除了向 Output 写入日志外，还会向 LocalFsConfig.Path 写入日志，且按照日志级别写入不同的文件
	//In addition to writing logs to Output, it will also write logs to LocalFsConfig.Path, and write logs to different files according to the log level
	LocalFsConfig LocalFsConfig
	// RotateConfig 除了向 Output 写入日志外，还会向 RotateConfig.Path 写入日志，按照日志级别写入不同的文件，且按配置轮转或压缩日志
	// In addition to writing logs to Output, it will also write logs to RotateConfig.Path, and write logs to different files according to the log level, and rotate or compress logs according to the configuration
	RotateConfig RotateConfig
	Caller       bool
	Timestamp    bool
	// GetTraceId 获取 trace_id 的函数
	// GetTraceId is a function to get the trace_id
	// 默认实现使用 opentelemetry 的 trace_id
	GetTraceIdFunc func(ctx context.Context) string
	// OtlpEnabled 是否启用 otlp
	OtlpEnabled bool
}

type LocalFsConfig struct {
	Path string
	// ErrorFileLevel default is ErrorLevel
	ErrorFileLevel Level
	// ErrorFileExt default is add ".err", if you use 'a/b/c.log', the error file is 'a/b/c.err.log'
	ErrorFileExt string
}

type RotateConfig struct {
	Path               string
	MaxSizeKB          int
	MaxRetainFileCount int
	MaxAge             time.Duration
	Compress           bool
	// ErrorFileLevel default is ErrorLevel
	ErrorFileLevel Level
	// ErrorFileExt default is add ".err", if you use 'a/b/c.log', the error file is 'a/b/c.err.log'
	ErrorFileExt string
}

var DefaultGetTraceIdFunc = func(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

func defaultOptions() *Options {
	return &Options{
		Type:           LoggerTypeZap,
		Level:          LevelInfo,
		Caller:         true,
		Timestamp:      true,
		Output:         os.Stdout,
		GetTraceIdFunc: DefaultGetTraceIdFunc,
	}
}

type Option func(*Options)

func WithOtlpEnabled(enabled bool) Option {
	return func(o *Options) {
		o.OtlpEnabled = enabled
	}
}

func WithOutput(w io.Writer) Option {
	return func(o *Options) {
		o.Output = w
	}
}

func WithLoggerType(t LoggerType) Option {
	return func(o *Options) {
		o.Type = t
	}
}

func WithLocalFs(config LocalFsConfig) Option {

	if config.ErrorFileLevel == "" {
		config.ErrorFileLevel = LevelError
	}

	if config.ErrorFileExt == "" {
		config.ErrorFileExt = ".err"
	}

	return func(o *Options) {
		o.LocalFsConfig = config
	}
}
func WithRotate(config RotateConfig) Option {
	if config.ErrorFileLevel == "" {
		config.ErrorFileLevel = LevelError
	}

	if config.ErrorFileExt == "" {
		config.ErrorFileExt = ".err"
	}

	return func(o *Options) {
		o.RotateConfig = config
	}
}

func WithCaller() Option {
	return func(o *Options) {
		o.Caller = true
	}
}

func WithTimestamp() Option {
	return func(o *Options) {
		o.Timestamp = true
	}
}

func WithLevel(level Level) Option {
	return func(o *Options) {
		o.Level = level
	}
}

func WithServiceName(serviceName string) Option {
	return func(o *Options) {
		o.ServiceName = serviceName
	}
}
