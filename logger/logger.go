package logger

import (
	"context"
	"fmt"
)

type Level string

func (l Level) String() string {
	return string(l)
}

const (
	LevelTrace Level = "trace"
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
	LevelFatal Level = "fatal"
	LevelPanic Level = "panic"
)

type LoggerType string

const (
	LoggerTypeLogrus  LoggerType = "logrus"
	LoggerTypeZap     LoggerType = "zap"
	LoggerTypeSlog    LoggerType = "slog"
	LoggerTypeZerolog LoggerType = "zerolog"
)

const FieldKeyTimestamp = "timestamp"
const FieldKeyServiceName = "service"
const FieldKeyCaller = "caller"
const FieldKeyTime = "time"
const FieldKeyMsg = "msg"
const FieldKeyLevel = "level"
const FieldKeyError = "error"
const FieldKeyTraceId = "trace_id"

type Logger interface {
	WithError(err error) Logger
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	Log(ctx context.Context, level Level, args ...interface{})
	Trace(ctx context.Context, args ...interface{})
	Debug(ctx context.Context, args ...interface{})
	Info(ctx context.Context, args ...interface{})
	Warn(ctx context.Context, args ...interface{})
	Warning(ctx context.Context, args ...interface{})
	Print(ctx context.Context, args ...interface{})
	Error(ctx context.Context, args ...interface{})
	Panic(ctx context.Context, args ...interface{})
	Fatal(ctx context.Context, args ...interface{})
	Logf(ctx context.Context, level Level, format string, args ...interface{})
	Tracef(ctx context.Context, format string, args ...interface{})
	Debugf(ctx context.Context, format string, args ...interface{})
	Infof(ctx context.Context, format string, args ...interface{})
	Warnf(ctx context.Context, format string, args ...interface{})
	Warningf(ctx context.Context, format string, args ...interface{})
	Printf(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
	Panicf(ctx context.Context, format string, args ...interface{})
	Fatalf(ctx context.Context, format string, args ...interface{})
	Logln(ctx context.Context, level Level, args ...interface{})
	Traceln(ctx context.Context, args ...interface{})
	Debugln(ctx context.Context, args ...interface{})
	Infoln(ctx context.Context, args ...interface{})
	Warnln(ctx context.Context, args ...interface{})
	Warningln(ctx context.Context, args ...interface{})
	Println(ctx context.Context, args ...interface{})
	Errorln(ctx context.Context, args ...interface{})
	Panicln(ctx context.Context, args ...interface{})
	Fatalln(ctx context.Context, args ...interface{})
}

var logger Logger = New()

func ResetLogger(l Logger) {
	logger = l
}

func ResetLoggerWithOptions(opts ...Option) Logger {
	logger = New(opts...)
	return logger
}

func New(opts ...Option) Logger {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	var l Logger
	switch options.Type {
	case LoggerTypeLogrus:
		l = newLogrusLogger(*options)
	case LoggerTypeZerolog:
		l = newDefaultZerologLogger(*options)
	case LoggerTypeZap:
		l = newZapLogger(*options)
	case LoggerTypeSlog:
		l = newSlogLogger(*options)
	default:
		panic(fmt.Sprintf("unsupported logger type: %s", options.Type))
	}

	return l
}

func WithError(err error) Logger {
	return logger.WithError(err)
}

func WithField(key string, value interface{}) Logger {
	return logger.WithField(key, value)
}

func WithFields(fields map[string]interface{}) Logger {
	return logger.WithFields(fields)
}

func Log(ctx context.Context, level Level, args ...interface{}) {
	logger.Log(ctx, level, args...)
}

func Trace(ctx context.Context, args ...interface{}) {
	logger.Trace(ctx, args...)
}

func Debug(ctx context.Context, args ...interface{}) {
	logger.Debug(ctx, args...)
}

func Info(ctx context.Context, args ...interface{}) {
	logger.Info(ctx, args...)
}

func Warn(ctx context.Context, args ...interface{}) {
	logger.Warn(ctx, args...)
}

func Warning(ctx context.Context, args ...interface{}) {
	logger.Warning(ctx, args...)
}

func Print(ctx context.Context, args ...interface{}) {
	logger.Print(ctx, args...)
}

func Error(ctx context.Context, args ...interface{}) {
	logger.Error(ctx, args...)
}

func Panic(ctx context.Context, args ...interface{}) {
	logger.Panic(ctx, args...)
}

func Fatal(ctx context.Context, args ...interface{}) {
	logger.Fatal(ctx, args...)
}

func Logf(ctx context.Context, level Level, format string, args ...interface{}) {
	logger.Logf(ctx, level, format, args...)
}

func Tracef(ctx context.Context, format string, args ...interface{}) {
	logger.Tracef(ctx, format, args...)
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	logger.Debugf(ctx, format, args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	logger.Infof(ctx, format, args...)
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	logger.Warnf(ctx, format, args...)
}

func Warningf(ctx context.Context, format string, args ...interface{}) {
	logger.Warningf(ctx, format, args...)
}

func Printf(ctx context.Context, format string, args ...interface{}) {
	logger.Printf(ctx, format, args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	logger.Errorf(ctx, format, args...)
}

func Panicf(ctx context.Context, format string, args ...interface{}) {
	logger.Panicf(ctx, format, args...)
}

func Fatalf(ctx context.Context, format string, args ...interface{}) {
	logger.Fatalf(ctx, format, args...)
}

func Logln(ctx context.Context, level Level, args ...interface{}) {
	logger.Logln(ctx, level, args...)
}

func Traceln(ctx context.Context, args ...interface{}) {
	logger.Traceln(ctx, args...)
}

func Debugln(ctx context.Context, args ...interface{}) {
	logger.Debugln(ctx, args...)
}

func Infoln(ctx context.Context, args ...interface{}) {
	logger.Infoln(ctx, args...)
}

func Warnln(ctx context.Context, args ...interface{}) {
	logger.Warnln(ctx, args...)
}

func Warningln(ctx context.Context, args ...interface{}) {
	logger.Warningln(ctx, args...)
}

func Println(ctx context.Context, args ...interface{}) {
	logger.Println(ctx, args...)
}

func Errorln(ctx context.Context, args ...interface{}) {
	logger.Errorln(ctx, args...)
}

func Panicln(ctx context.Context, args ...interface{}) {
	logger.Panicln(ctx, args...)
}

func Fatalln(ctx context.Context, args ...interface{}) {
	logger.Fatalln(ctx, args...)
}
