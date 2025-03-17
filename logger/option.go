package logger

import (
	"io"
	"os"
	"time"
)

type LoggerType string

const (
	LoggerTypeLogrus  LoggerType = "logrus"
	LoggerTypeZap     LoggerType = "zap"
	LoggerTypeSlog    LoggerType = "slog"
	LoggerTypeZerolog LoggerType = "zerolog"
)

type Options struct {
	Type         LoggerType
	Level        Level
	ServiceName  string
	LocalFsPath  string
	RotateConfig RotateConfig
	Caller       bool
	CallerSkip   int
	Timestamp    bool
	Output       io.Writer
}

type RotateConfig struct {
	Path       string
	RotateTime time.Duration
	ExpireTime time.Duration
}

func defaultOptions() *Options {
	return &Options{
		Type:       LoggerTypeZerolog,
		Level:      LevelInfo,
		Caller:     true,
		CallerSkip: 0,
		Timestamp:  true,
		Output:     os.Stdout,
	}
}

type Option func(*Options)

func WithCallerSkip(skip int) Option {
	return func(o *Options) {
		o.CallerSkip = skip
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

func WithLocalFsPath(path string) Option {
	return func(o *Options) {
		o.LocalFsPath = path
	}
}
func WithRotate(path string, config RotateConfig) Option {
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
