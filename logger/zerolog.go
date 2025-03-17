package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

type zerologLogger struct {
	Logger zerolog.Logger
}

var _ Logger = &zerologLogger{}

func newZerologLogger(opt *Options) *zerologLogger {

	zerolog.MessageFieldName = FieldKeyMsg
	zerolog.CallerFieldName = FieldKeyCaller
	zerolog.TimestampFieldName = FieldKeyTime
	zerolog.TimeFieldFormat = time.DateTime
	logger := zerolog.New(opt.Output)

	if opt.Caller {
		logger = logger.With().CallerWithSkipFrameCount(3 + opt.CallerSkip).Logger()
	}

	if opt.Level != "" {
		logger = logger.Level(levelToZerologLevel(opt.Level))
	}

	if opt.ServiceName != "" {
		logger = logger.Hook(newZerologServiceHook(opt.ServiceName))
	}

	if opt.Timestamp {
		logger = logger.Hook(newZerologTimestampHook())
	}

	return &zerologLogger{
		Logger: logger,
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
	newLogger := l.Logger.With().Err(err).Logger()
	return &zerologLogger{Logger: newLogger}
}

func (l *zerologLogger) WithField(key string, value interface{}) Logger {
	newLogger := l.Logger.With().Interface(key, value).Logger()
	return &zerologLogger{Logger: newLogger}
}

func (l *zerologLogger) WithFields(fields map[string]interface{}) Logger {
	ctx := l.Logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &zerologLogger{Logger: ctx.Logger()}
}

func (l *zerologLogger) Log(ctx context.Context, level Level, args ...interface{}) {
	l.Logger.Log().Str(FieldKeyLevel, level.String()).Timestamp().Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Trace(ctx context.Context, args ...interface{}) {
	l.Logger.Trace().Timestamp().Ctx(ctx).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Debug(ctx context.Context, args ...interface{}) {
	l.Logger.Debug().Timestamp().Ctx(ctx).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Info(ctx context.Context, args ...interface{}) {
	l.Logger.Info().Timestamp().Ctx(ctx).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Warn(ctx context.Context, args ...interface{}) {
	l.Logger.Warn().Timestamp().Ctx(ctx).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Warning(ctx context.Context, args ...interface{}) {
	l.Logger.Warn().Timestamp().Ctx(ctx).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Print(ctx context.Context, args ...interface{}) {
	l.Info(ctx, args...)
}

func (l *zerologLogger) Error(ctx context.Context, args ...interface{}) {
	l.Logger.Error().Timestamp().Ctx(ctx).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Panic(ctx context.Context, args ...interface{}) {
	l.Logger.Panic().Timestamp().Ctx(ctx).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Fatal(ctx context.Context, args ...interface{}) {
	l.Logger.Fatal().Timestamp().Ctx(ctx).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Logf(ctx context.Context, level Level, format string, args ...interface{}) {
	l.Logger.Log().Timestamp().Ctx(ctx).Str(FieldKeyLevel, level.String()).Msgf(format, args...)
}

func (l *zerologLogger) Tracef(ctx context.Context, format string, args ...interface{}) {
	l.Logger.Trace().Timestamp().Ctx(ctx).Msgf(format, args...)
}

func (l *zerologLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.Logger.Debug().Timestamp().Ctx(ctx).Msgf(format, args...)
}

func (l *zerologLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.Logger.Info().Timestamp().Ctx(ctx).Msgf(format, args...)
}

func (l *zerologLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.Logger.Warn().Timestamp().Ctx(ctx).Msgf(format, args...)
}

func (l *zerologLogger) Warningf(ctx context.Context, format string, args ...interface{}) {
	l.Logger.Warn().Timestamp().Ctx(ctx).Msgf(format, args...)
}

func (l *zerologLogger) Printf(ctx context.Context, format string, args ...interface{}) {
	l.Infof(ctx, format, args...)
}

func (l *zerologLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.Logger.Error().Timestamp().Ctx(ctx).Msgf(format, args...)
}

func (l *zerologLogger) Panicf(ctx context.Context, format string, args ...interface{}) {
	l.Logger.Panic().Timestamp().Ctx(ctx).Msgf(format, args...)
}

func (l *zerologLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	l.Logger.Fatal().Timestamp().Ctx(ctx).Msgf(format, args...)
}

func (l *zerologLogger) Logln(ctx context.Context, level Level, args ...interface{}) {
	l.Logger.Log().Timestamp().Ctx(ctx).Str(FieldKeyLevel, level.String()).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Traceln(ctx context.Context, args ...interface{}) {
	l.Logger.Trace().Timestamp().Ctx(ctx).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Debugln(ctx context.Context, args ...interface{}) {
	l.Logger.Debug().Timestamp().Ctx(ctx).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Infoln(ctx context.Context, args ...interface{}) {
	l.Logger.Info().Timestamp().Ctx(ctx).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Warnln(ctx context.Context, args ...interface{}) {
	l.Logger.Warn().Timestamp().Ctx(ctx).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Warningln(ctx context.Context, args ...interface{}) {
	l.Logger.Warn().Timestamp().Ctx(ctx).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Println(ctx context.Context, args ...interface{}) {
	l.Infoln(ctx, args...)
}

func (l *zerologLogger) Errorln(ctx context.Context, args ...interface{}) {
	l.Logger.Error().Timestamp().Ctx(ctx).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Panicln(ctx context.Context, args ...interface{}) {
	l.Logger.Panic().Timestamp().Ctx(ctx).Msg(fmt.Sprint(args...))
}

func (l *zerologLogger) Fatalln(ctx context.Context, args ...interface{}) {
	l.Logger.Fatal().Timestamp().Ctx(ctx).Msg(fmt.Sprint(args...))
}
