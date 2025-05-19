package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	Logger *logrus.Logger
	Entry  *logrus.Entry
	Opt    Options
}

var _ Logger = &logrusLogger{}

func newLogrusLogger(opt Options) *logrusLogger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.DateTime,
	})

	if opt.Level != "" {
		logger.SetLevel(levelToLogrusLevel(opt.Level))
	}

	if opt.Output != nil {
		logger.SetOutput(opt.Output)
	}

	if opt.Caller {
		logger.SetReportCaller(false)
		logger.AddHook(newLogrusCallerHook())
	}

	if opt.Timestamp {
		logger.AddHook(newLogrusTimestampHook())
	}

	if opt.ServiceName != "" {
		logger.AddHook(newLogrusServiceHook(opt.ServiceName))
	}

	if opt.GetTraceIdFunc != nil {
		logger.AddHook(newLogrusTraceIdHook(opt.GetTraceIdFunc))
	}

	if opt.LocalFsConfig.Path != "" {
		logger.AddHook(newLogrusLocalFsHook(opt.LocalFsConfig))
	}

	if opt.RotateConfig.Path != "" {
		hook, err := newLogrusRotateHook(opt.RotateConfig)
		if err != nil {
			panic(fmt.Sprintf("new logrus rotate hook error: %s", err))
		}
		logger.AddHook(hook)
	}

	if opt.OtlpEnabled {
		logger.AddHook(newLogrusOtlpHook())
	}

	return &logrusLogger{
		Logger: logger,
		Entry:  logrus.NewEntry(logger),
		Opt:    opt,
	}
}

func levelToLogrusLevel(level Level) logrus.Level {
	switch level {
	case LevelDebug:
		return logrus.DebugLevel
	case LevelInfo:
		return logrus.InfoLevel
	case LevelWarn:
		return logrus.WarnLevel
	case LevelError:
		return logrus.ErrorLevel
	case LevelFatal:
		return logrus.FatalLevel
	case LevelPanic:
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

func (l *logrusLogger) WithError(err error) Logger {
	return &logrusLogger{
		Logger: l.Logger,
		Entry:  l.Entry.WithError(err),
	}
}

func (l *logrusLogger) WithField(key string, value interface{}) Logger {
	return &logrusLogger{
		Logger: l.Logger,
		Entry:  l.Entry.WithField(key, value),
	}
}

func (l *logrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &logrusLogger{
		Logger: l.Logger,
		Entry:  l.Entry.WithFields(fields),
	}
}

func (l *logrusLogger) Log(ctx context.Context, level Level, args ...interface{}) {
	l.Entry.WithContext(ctx).Log(levelToLogrusLevel(level), args...)
}

func (l *logrusLogger) Trace(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelTrace, args...)
}

func (l *logrusLogger) Debug(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelDebug, args...)
}

func (l *logrusLogger) Info(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelInfo, args...)
}

func (l *logrusLogger) Warn(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelWarn, args...)
}

func (l *logrusLogger) Warning(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelWarn, args...)
}

func (l *logrusLogger) Print(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelInfo, args...)
}

func (l *logrusLogger) Error(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelError, args...)
}

func (l *logrusLogger) Panic(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelPanic, args...)
}

func (l *logrusLogger) Fatal(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelFatal, args...)
}

func (l *logrusLogger) Logf(ctx context.Context, level Level, format string, args ...interface{}) {
	l.Log(ctx, level, fmt.Sprintf(format, args...))
}

func (l *logrusLogger) Tracef(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelTrace, format, args...)
}

func (l *logrusLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelDebug, format, args...)
}

func (l *logrusLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelInfo, format, args...)
}

func (l *logrusLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelWarn, format, args...)
}

func (l *logrusLogger) Warningf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelWarn, format, args...)
}

func (l *logrusLogger) Printf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelInfo, format, args...)
}

func (l *logrusLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelError, format, args...)
}

func (l *logrusLogger) Panicf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelPanic, format, args...)
}

func (l *logrusLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, LevelFatal, format, args...)
}

func (l *logrusLogger) Logln(ctx context.Context, level Level, args ...interface{}) {
	l.Log(ctx, level, fmt.Sprintln(args...))
}

func (l *logrusLogger) Traceln(ctx context.Context, args ...interface{}) {
	l.Logln(ctx, LevelTrace, args...)
}

func (l *logrusLogger) Debugln(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelDebug, args...)
}

func (l *logrusLogger) Infoln(ctx context.Context, args ...interface{}) {
	l.Logln(ctx, LevelInfo, args...)
}

func (l *logrusLogger) Warnln(ctx context.Context, args ...interface{}) {
	l.Logln(ctx, LevelWarn, args...)
}

func (l *logrusLogger) Warningln(ctx context.Context, args ...interface{}) {
	l.Logln(ctx, LevelWarn, args...)
}

func (l *logrusLogger) Println(ctx context.Context, args ...interface{}) {
	l.Logln(ctx, LevelInfo, args...)
}

func (l *logrusLogger) Errorln(ctx context.Context, args ...interface{}) {
	l.Logln(ctx, LevelError, args...)
}

func (l *logrusLogger) Panicln(ctx context.Context, args ...interface{}) {
	l.Logln(ctx, LevelPanic, args...)
}

func (l *logrusLogger) Fatalln(ctx context.Context, args ...interface{}) {
	l.Logln(ctx, LevelFatal, args...)
}
