package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

type zerologLogger struct {
	Logger zerolog.Logger
	Opt    *Options
	Err    error
	Fields map[string]interface{}
}

var _ Logger = &zerologLogger{}

func newZerologLogger(logger zerolog.Logger, opt *Options) *zerologLogger {
	zerolog.MessageFieldName = FieldKeyMsg
	zerolog.CallerFieldName = FieldKeyCaller
	zerolog.TimestampFieldName = FieldKeyTime
	zerolog.TimeFieldFormat = time.DateTime

	if opt.Level != "" {
		logger = logger.Level(levelToZerologLevel(opt.Level))
	}

	if opt.ServiceName != "" {
		logger = logger.Hook(newZerologServiceHook(opt.ServiceName))
	}

	if opt.Timestamp {
		logger = logger.Hook(newZerologTimestampHook())
	}

	if opt.LocalFsConfig.Path != "" {
		hook := newZerologLocalFsHook(logger, opt, opt.LocalFsConfig)
		logger = logger.Hook(hook)
	}

	if opt.RotateConfig.Path != "" {
		hook, err := newZerologRotateHook(logger, opt, opt.RotateConfig)
		if err != nil {
			panic(fmt.Sprintf("new zerolog rotate hook error: %s", err))
		}
		logger = logger.Hook(hook)
	}

	if opt.Caller {
		logger = logger.With().CallerWithSkipFrameCount(4 + opt.CallerSkip).Logger()
	}

	return &zerologLogger{
		Logger: logger,
		Opt:    opt,
		Err:    nil,
		Fields: make(map[string]interface{}),
	}
}

func levelToZerologLevel(level Level) zerolog.Level {
	switch level {
	case LevelTrace:
		return zerolog.TraceLevel
	case LevelDebug:
		return zerolog.DebugLevel
	case LevelInfo:
		return zerolog.InfoLevel
	case LevelWarn:
		return zerolog.WarnLevel
	case LevelError:
		return zerolog.ErrorLevel
	case LevelPanic:
		return zerolog.PanicLevel
	case LevelFatal:
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

func (l *zerologLogger) WithError(err error) Logger {
	return &zerologLogger{
		Logger: zerolog.New(l.Opt.Output),
		Opt:    l.Opt,
		Err:    err,
		Fields: l.Fields,
	}
}

func (l *zerologLogger) WithField(key string, value interface{}) Logger {
	return l.WithFields(map[string]interface{}{key: value})
}

func (l *zerologLogger) WithFields(fields map[string]interface{}) Logger {
	allFields := make(map[string]interface{})
	for k, v := range l.Fields {
		allFields[k] = v
	}
	for k, v := range fields {
		allFields[k] = v
	}

	return &zerologLogger{
		Logger: zerolog.New(l.Opt.Output),
		Opt:    l.Opt,
		Err:    l.Err,
		Fields: allFields,
	}
}

func (l *zerologLogger) Log(ctx context.Context, level Level, args ...interface{}) {
	loggerCtx := zerolog.New(l.Opt.Output).With()
	for k, v := range l.Fields {
		loggerCtx = loggerCtx.Interface(k, v)
	}

	if l.Err != nil {
		loggerCtx = loggerCtx.Err(l.Err)
	}

	logger := loggerCtx.Ctx(ctx).Timestamp().Logger()
	newLogger := newZerologLogger(logger, l.Opt)
	newLogger.Logger.WithLevel(levelToZerologLevel(level)).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Trace(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelTrace, args...)
}

func (l *zerologLogger) Debug(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelDebug, args...)
}

func (l *zerologLogger) Info(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelInfo, args...)
}

func (l *zerologLogger) Warn(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelWarn, args...)
}

func (l *zerologLogger) Warning(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelWarn, args...)
}

func (l *zerologLogger) Print(ctx context.Context, args ...interface{}) {
	l.Info(ctx, args...)
}

func (l *zerologLogger) Error(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelError, args...)
}

func (l *zerologLogger) Panic(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelPanic, args...)
}

func (l *zerologLogger) Fatal(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelFatal, args...)
}

func (l *zerologLogger) Logf(ctx context.Context, level Level, format string, args ...interface{}) {
	loggerCtx := zerolog.New(l.Opt.Output).With()
	for k, v := range l.Fields {
		loggerCtx = loggerCtx.Interface(k, v)
	}

	if l.Err != nil {
		loggerCtx = loggerCtx.Err(l.Err)
	}

	logger := loggerCtx.Ctx(ctx).Timestamp().Logger()
	newLogger := newZerologLogger(logger, l.Opt)
	newLogger.Logger.WithLevel(levelToZerologLevel(level)).Msgf(format, args...)
}

func (l *zerologLogger) Tracef(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelTrace, format, args...)
}

func (l *zerologLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelDebug, format, args...)
}

func (l *zerologLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelInfo, format, args...)
}

func (l *zerologLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelWarn, format, args...)
}

func (l *zerologLogger) Warningf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelWarn, format, args...)
}

func (l *zerologLogger) Printf(ctx context.Context, format string, args ...interface{}) {
	l.Infof(ctx, format, args...)
}

func (l *zerologLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelError, format, args...)
}

func (l *zerologLogger) Panicf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelPanic, format, args...)
}

func (l *zerologLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelFatal, format, args...)
}

func (l *zerologLogger) Logln(ctx context.Context, level Level, args ...interface{}) {
	loggerCtx := zerolog.New(l.Opt.Output).With()
	for k, v := range l.Fields {
		loggerCtx = loggerCtx.Interface(k, v)
	}

	if l.Err != nil {
		loggerCtx = loggerCtx.Err(l.Err)
	}

	logger := loggerCtx.Ctx(ctx).Timestamp().Logger()
	newLogger := newZerologLogger(logger, l.Opt)
	newLogger.Logger.WithLevel(levelToZerologLevel(level)).Msg(fmt.Sprintln(args...))
}

func (l *zerologLogger) Traceln(ctx context.Context, args ...interface{}) {
	l.Logln(ctx, LevelTrace, args...)
}

func (l *zerologLogger) Debugln(ctx context.Context, args ...interface{}) {
	l.Logln(ctx, LevelDebug, args...)
}

func (l *zerologLogger) Infoln(ctx context.Context, args ...interface{}) {
	l.Logln(ctx, LevelInfo, args...)
}

func (l *zerologLogger) Warnln(ctx context.Context, args ...interface{}) {
	l.Logln(ctx, LevelWarn, args...)
}

func (l *zerologLogger) Warningln(ctx context.Context, args ...interface{}) {
	l.Logln(ctx, LevelWarn, args...)
}

func (l *zerologLogger) Println(ctx context.Context, args ...interface{}) {
	l.Infoln(ctx, args...)
}

func (l *zerologLogger) Errorln(ctx context.Context, args ...interface{}) {
	l.Logln(ctx, LevelError, args...)
}

func (l *zerologLogger) Panicln(ctx context.Context, args ...interface{}) {
	l.Logln(ctx, LevelPanic, args...)
}

func (l *zerologLogger) Fatalln(ctx context.Context, args ...interface{}) {
	l.Logln(ctx, LevelFatal, args...)
}
