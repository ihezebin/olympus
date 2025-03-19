package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	Logger *zap.Logger
	Fields []zap.Field
}

var _ Logger = &zapLogger{}

func newZapLogger(opt Options) *zapLogger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = FieldKeyTime
	encoderConfig.MessageKey = FieldKeyMsg
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	encoderConfig.CallerKey = "" // 禁用 zap 的 caller

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 设置日志级别
	level := zap.NewAtomicLevelAt(levelToZapLevel(opt.Level))

	// 如果指定了 Output，使用自定义的 WriteSyncer
	var ws zapcore.WriteSyncer
	if opt.Output != nil {
		ws = zapcore.AddSync(opt.Output)
	} else {
		// 默认输出到标准输出
		ws = zapcore.AddSync(os.Stdout)
	}

	// 创建基础 core
	core := zapcore.NewCore(encoder, ws, level)
	hook := newZapHook(core, encoder, opt)
	logger := zap.New(hook)

	return &zapLogger{
		Logger: logger,
	}
}

func levelToZapLevel(level Level) zapcore.Level {
	switch level {
	case LevelDebug:
		return zapcore.DebugLevel
	case LevelInfo:
		return zapcore.InfoLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelError:
		return zapcore.ErrorLevel
	case LevelPanic:
		return zapcore.PanicLevel
	case LevelFatal:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func (l *zapLogger) withContext(ctx context.Context) *zap.Logger {
	fields := make([]zap.Field, 0)
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields = append(fields, zap.Any("trace_id", traceID))

	}
	l.Logger = l.Logger.With(fields...)
	return l.Logger
}

func (l *zapLogger) WithError(err error) Logger {
	newFields := make([]zap.Field, 0)
	newFields = append(newFields, l.Fields...)
	newFields = append(newFields, zap.Error(err))
	return &zapLogger{Logger: l.Logger, Fields: newFields}
}

func (l *zapLogger) WithField(key string, value interface{}) Logger {
	newFields := make([]zap.Field, 0)
	newFields = append(newFields, l.Fields...)
	newFields = append(newFields, zap.Any(key, value))
	return &zapLogger{Logger: l.Logger, Fields: newFields}
}

func (l *zapLogger) WithFields(fields map[string]interface{}) Logger {
	newFields := make([]zap.Field, 0, len(fields))
	newFields = append(newFields, l.Fields...)
	for k, v := range fields {
		newFields = append(newFields, zap.Any(k, v))
	}
	return &zapLogger{Logger: l.Logger, Fields: newFields}
}

func (l *zapLogger) Log(ctx context.Context, level Level, args ...interface{}) {
	l.withContext(ctx).Log(levelToZapLevel(level), fmt.Sprint(args...), l.Fields...)
}

func (l *zapLogger) Trace(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelTrace, args...)
}

func (l *zapLogger) Debug(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelDebug, args...)
}

func (l *zapLogger) Info(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelInfo, args...)
}

func (l *zapLogger) Warn(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelWarn, args...)
}

func (l *zapLogger) Warning(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelWarn, args...)
}

func (l *zapLogger) Print(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelInfo, args...)
}

func (l *zapLogger) Error(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelError, args...)
}

func (l *zapLogger) Panic(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelPanic, args...)
}

func (l *zapLogger) Fatal(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelFatal, args...)
}

func (l *zapLogger) Logf(ctx context.Context, level Level, format string, args ...interface{}) {
	l.Log(ctx, level, fmt.Sprintf(format, args...))
}

func (l *zapLogger) Tracef(ctx context.Context, format string, args ...interface{}) {
	l.Log(ctx, LevelTrace, fmt.Sprintf(format, args...))
}

func (l *zapLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.Log(ctx, LevelDebug, fmt.Sprintf(format, args...))
}

func (l *zapLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.Log(ctx, LevelInfo, fmt.Sprintf(format, args...))
}

func (l *zapLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.Log(ctx, LevelWarn, fmt.Sprintf(format, args...))
}

func (l *zapLogger) Warningf(ctx context.Context, format string, args ...interface{}) {
	l.Log(ctx, LevelWarn, fmt.Sprintf(format, args...))
}

func (l *zapLogger) Printf(ctx context.Context, format string, args ...interface{}) {
	l.Log(ctx, LevelInfo, fmt.Sprintf(format, args...))
}

func (l *zapLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.Log(ctx, LevelError, fmt.Sprintf(format, args...))
}

func (l *zapLogger) Panicf(ctx context.Context, format string, args ...interface{}) {
	l.Log(ctx, LevelPanic, fmt.Sprintf(format, args...))
}

func (l *zapLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	l.Log(ctx, LevelFatal, fmt.Sprintf(format, args...))
}

func (l *zapLogger) Logln(ctx context.Context, level Level, args ...interface{}) {
	l.Log(ctx, level, fmt.Sprintln(args...))
}

func (l *zapLogger) Traceln(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelTrace, fmt.Sprintln(args...))
}

func (l *zapLogger) Debugln(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelDebug, fmt.Sprintln(args...))
}

func (l *zapLogger) Infoln(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelInfo, fmt.Sprintln(args...))
}

func (l *zapLogger) Warnln(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelWarn, fmt.Sprintln(args...))
}

func (l *zapLogger) Warningln(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelWarn, fmt.Sprintln(args...))
}

func (l *zapLogger) Println(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelInfo, fmt.Sprintln(args...))
}

func (l *zapLogger) Errorln(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelError, fmt.Sprintln(args...))
}

func (l *zapLogger) Panicln(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelPanic, fmt.Sprintln(args...))
}

func (l *zapLogger) Fatalln(ctx context.Context, args ...interface{}) {
	l.Log(ctx, LevelFatal, fmt.Sprintln(args...))
}
