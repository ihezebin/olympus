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
}

var _ Logger = &logrusLogger{}

func newLogrusLogger(opt *Options) *logrusLogger {
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
		logger.AddHook(newLogrusCallerHook(8 + opt.CallerSkip))
	}

	if opt.Timestamp {
		logger.AddHook(newLogrusTimestampHook())
	}

	if opt.ServiceName != "" {
		logger.AddHook(newLogrusServiceHook(opt.ServiceName))
	}

	if opt.LocalFsConfig.Path != "" {
		hook, err := newLogrusLocalFsHook(opt.LocalFsConfig)
		if err != nil {
			panic(fmt.Sprintf("new logrus local fs hook error: %s", err))
		}
		logger.AddHook(hook)
	}

	if opt.RotateConfig.Path != "" {
		hook, err := newLogrusRotateHook(opt.RotateConfig)
		if err != nil {
			panic(fmt.Sprintf("new logrus rotate hook error: %s", err))
		}
		logger.AddHook(hook)
	}

	return &logrusLogger{
		Logger: logger,
		Entry:  logrus.NewEntry(logger),
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
	l.Entry = l.Entry.WithError(err)
	return l
}

func (l *logrusLogger) WithField(key string, value interface{}) Logger {
	l.Entry = l.Entry.WithField(key, value)
	return l
}

func (l *logrusLogger) WithFields(fields map[string]interface{}) Logger {
	l.Entry = l.Entry.WithFields(fields)
	return l
}

func (l *logrusLogger) Log(ctx context.Context, level Level, args ...interface{}) {
	l.Entry.WithContext(ctx).Log(levelToLogrusLevel(level), args...)
}

func (l *logrusLogger) Trace(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Trace(args...)
}

func (l *logrusLogger) Debug(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Debug(args...)
}

func (l *logrusLogger) Info(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Info(args...)
}

func (l *logrusLogger) Warn(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Warn(args...)
}

func (l *logrusLogger) Warning(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Warning(args...)
}

func (l *logrusLogger) Print(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Print(args...)
}

func (l *logrusLogger) Error(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Error(args...)
}

func (l *logrusLogger) Panic(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Panic(args...)
}

func (l *logrusLogger) Fatal(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Fatal(args...)
}

func (l *logrusLogger) Logf(ctx context.Context, level Level, format string, args ...interface{}) {
	l.Entry.WithContext(ctx).Logf(levelToLogrusLevel(level), format, args...)
}

func (l *logrusLogger) Tracef(ctx context.Context, format string, args ...interface{}) {
	l.Entry.WithContext(ctx).Tracef(format, args...)
}

func (l *logrusLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.Entry.WithContext(ctx).Debugf(format, args...)
}

func (l *logrusLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.Entry.WithContext(ctx).Infof(format, args...)
}

func (l *logrusLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.Entry.WithContext(ctx).Warnf(format, args...)
}

func (l *logrusLogger) Warningf(ctx context.Context, format string, args ...interface{}) {
	l.Entry.WithContext(ctx).Warningf(format, args...)
}

func (l *logrusLogger) Printf(ctx context.Context, format string, args ...interface{}) {
	l.Entry.WithContext(ctx).Printf(format, args...)
}

func (l *logrusLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.Entry.WithContext(ctx).Errorf(format, args...)
}

func (l *logrusLogger) Panicf(ctx context.Context, format string, args ...interface{}) {
	l.Entry.WithContext(ctx).Panicf(format, args...)
}

func (l *logrusLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	l.Entry.WithContext(ctx).Fatalf(format, args...)
}

func (l *logrusLogger) Logln(ctx context.Context, level Level, args ...interface{}) {
	l.Entry.WithContext(ctx).Logln(levelToLogrusLevel(level), args...)
}

func (l *logrusLogger) Traceln(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Traceln(args...)
}

func (l *logrusLogger) Debugln(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Debugln(args...)
}

func (l *logrusLogger) Infoln(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Infoln(args...)
}

func (l *logrusLogger) Warnln(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Warnln(args...)
}

func (l *logrusLogger) Warningln(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Warningln(args...)
}

func (l *logrusLogger) Println(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Println(args...)
}

func (l *logrusLogger) Errorln(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Errorln(args...)
}

func (l *logrusLogger) Panicln(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Panicln(args...)
}

func (l *logrusLogger) Fatalln(ctx context.Context, args ...interface{}) {
	l.Entry.WithContext(ctx).Fatalln(args...)
}
