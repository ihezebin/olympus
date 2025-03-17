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
	Type          LoggerType
	Level         Level
	ServiceName   string
	LocalFsConfig LocalFsConfig
	RotateConfig  RotateConfig
	Caller        bool
	CallerSkip    int
	Timestamp     bool
	Output        io.Writer
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
