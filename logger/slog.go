package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
)

type slogLogger struct {
	Logger    *slog.Logger
	Timestamp bool
}

var _ Logger = &slogLogger{}

func newSlogLogger(opt *Options) *slogLogger {
	var handler slog.Handler

	handlerOpts := &slog.HandlerOptions{
		Level:     levelToSlogLevel(opt.Level),
		AddSource: opt.Caller,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.AnyValue(a.Value.Time().Format(time.DateTime))
			}

			// level 小写
			if a.Key == slog.LevelKey {
				a.Value = slog.StringValue(strings.ToLower(a.Value.String()))
			}

			// 添加 caller
			if a.Key == slog.SourceKey {
				a.Key = FieldKeyCaller
				caller := ""
				_, file, line, ok := runtime.Caller(7 + opt.CallerSkip)
				if ok {
					caller = fmt.Sprintf("%s:%d", file, line)
				}

				a.Value = slog.StringValue(caller)
			}

			return a
		},
	}

	handler = slog.NewJSONHandler(opt.Output, handlerOpts)

	if opt.ServiceName != "" {
		handler = handler.WithAttrs([]slog.Attr{
			slog.String(FieldKeyServiceName, opt.ServiceName),
		})
	}

	logger := slog.New(handler)

	return &slogLogger{
		Logger:    logger,
		Timestamp: opt.Timestamp,
	}
}

func levelToSlogLevel(level Level) slog.Level {
	switch level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	case LevelPanic:
		return slog.LevelError
	case LevelFatal:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (l *slogLogger) withTimestamp() *slogLogger {
	if !l.Timestamp {
		return l
	}
	newLogger := l.Logger.With(slog.Any(FieldKeyTimestamp, time.Now().Unix()))
	return &slogLogger{Logger: newLogger}
}

func (l *slogLogger) WithError(err error) Logger {
	newLogger := l.Logger.With(slog.Any(FieldKeyError, err))
	return &slogLogger{Logger: newLogger}
}

func (l *slogLogger) WithField(key string, value interface{}) Logger {
	newLogger := l.Logger.With(slog.Any(key, value))
	return &slogLogger{Logger: newLogger}
}

func (l *slogLogger) WithFields(fields map[string]interface{}) Logger {
	attrs := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		attrs = append(attrs, k, v)
	}
	newLogger := l.Logger.With(attrs...)
	return &slogLogger{Logger: newLogger}
}

func (l *slogLogger) Log(ctx context.Context, level Level, args ...interface{}) {
	l.withTimestamp().Logger.Log(ctx, levelToSlogLevel(level), fmt.Sprint(args...))
}

func (l *slogLogger) Trace(ctx context.Context, args ...interface{}) {
	l.withTimestamp().Logger.DebugContext(ctx, fmt.Sprint(args...))
}

func (l *slogLogger) Debug(ctx context.Context, args ...interface{}) {
	l.withTimestamp().Logger.DebugContext(ctx, fmt.Sprint(args...))
}

func (l *slogLogger) Info(ctx context.Context, args ...interface{}) {
	l.withTimestamp().Logger.InfoContext(ctx, fmt.Sprint(args...))
}

func (l *slogLogger) Warn(ctx context.Context, args ...interface{}) {
	l.withTimestamp().Logger.WarnContext(ctx, fmt.Sprint(args...))
}

func (l *slogLogger) Warning(ctx context.Context, args ...interface{}) {
	l.withTimestamp().Logger.WarnContext(ctx, fmt.Sprint(args...))
}

func (l *slogLogger) Print(ctx context.Context, args ...interface{}) {
	l.withTimestamp().Logger.InfoContext(ctx, fmt.Sprint(args...))
}

func (l *slogLogger) Error(ctx context.Context, args ...interface{}) {
	l.withTimestamp().Logger.ErrorContext(ctx, fmt.Sprint(args...))
}

func (l *slogLogger) Panic(ctx context.Context, args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.withTimestamp().Logger.Log(ctx, slog.LevelError, msg)
	panic(msg)
}

func (l *slogLogger) Fatal(ctx context.Context, args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.withTimestamp().Logger.Log(ctx, slog.LevelError, msg)
	os.Exit(1)
}

func (l *slogLogger) Logf(ctx context.Context, level Level, format string, args ...interface{}) {
	l.withTimestamp().Logger.Log(ctx, levelToSlogLevel(level), fmt.Sprintf(format, args...))
}

func (l *slogLogger) Tracef(ctx context.Context, format string, args ...interface{}) {
	l.withTimestamp().Logger.DebugContext(ctx, fmt.Sprintf(format, args...))
}

func (l *slogLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.withTimestamp().Logger.DebugContext(ctx, fmt.Sprintf(format, args...))
}

func (l *slogLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.withTimestamp().Logger.InfoContext(ctx, fmt.Sprintf(format, args...))
}

func (l *slogLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.withTimestamp().Logger.WarnContext(ctx, fmt.Sprintf(format, args...))
}

func (l *slogLogger) Warningf(ctx context.Context, format string, args ...interface{}) {
	l.withTimestamp().Logger.WarnContext(ctx, fmt.Sprintf(format, args...))
}

func (l *slogLogger) Printf(ctx context.Context, format string, args ...interface{}) {
	l.withTimestamp().Logger.InfoContext(ctx, fmt.Sprintf(format, args...))
}

func (l *slogLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.withTimestamp().Logger.ErrorContext(ctx, fmt.Sprintf(format, args...))
}

func (l *slogLogger) Panicf(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.withTimestamp().Logger.Log(ctx, slog.LevelError, msg)
	panic(msg)
}

func (l *slogLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.withTimestamp().Logger.Log(ctx, slog.LevelError, msg)
	os.Exit(1)
}

func (l *slogLogger) Logln(ctx context.Context, level Level, args ...interface{}) {
	l.withTimestamp().Logger.Log(ctx, levelToSlogLevel(level), fmt.Sprint(args...))
}

func (l *slogLogger) Traceln(ctx context.Context, args ...interface{}) {
	l.withTimestamp().Logger.DebugContext(ctx, fmt.Sprint(args...))
}

func (l *slogLogger) Debugln(ctx context.Context, args ...interface{}) {
	l.withTimestamp().Logger.DebugContext(ctx, fmt.Sprint(args...))
}

func (l *slogLogger) Infoln(ctx context.Context, args ...interface{}) {
	l.withTimestamp().Logger.InfoContext(ctx, fmt.Sprint(args...))
}

func (l *slogLogger) Warnln(ctx context.Context, args ...interface{}) {
	l.withTimestamp().Logger.WarnContext(ctx, fmt.Sprint(args...))
}

func (l *slogLogger) Warningln(ctx context.Context, args ...interface{}) {
	l.withTimestamp().Logger.WarnContext(ctx, fmt.Sprint(args...))
}

func (l *slogLogger) Println(ctx context.Context, args ...interface{}) {
	l.withTimestamp().Logger.InfoContext(ctx, fmt.Sprint(args...))
}

func (l *slogLogger) Errorln(ctx context.Context, args ...interface{}) {
	l.withTimestamp().Logger.ErrorContext(ctx, fmt.Sprint(args...))
}

func (l *slogLogger) Panicln(ctx context.Context, args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.withTimestamp().Logger.Log(ctx, slog.LevelError, msg)
	panic(msg)
}

func (l *slogLogger) Fatalln(ctx context.Context, args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.withTimestamp().Logger.Log(ctx, slog.LevelError, msg)
	os.Exit(1)
}
