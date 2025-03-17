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
}

var _ Logger = &zapLogger{}

func newZapLogger(opt *Options) *zapLogger {
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

	// 创建一个包装链，每个 hook 都包装前一个 core
	if opt.Timestamp {
		core = newZapTimestampHook(core)
	}
	if opt.Caller {
		core = newZapCallerHook(core, 6+opt.CallerSkip)
	}
	if opt.ServiceName != "" {
		core = newZapServiceHook(core, opt.ServiceName)
	}

	core = zapcore.NewTee(core)
	logger := zap.New(core,
		zap.AddStacktrace(zapcore.ErrorLevel),    // 添加错误堆栈
		zap.ErrorOutput(zapcore.Lock(os.Stderr)), // 添加错误输出
	)

	// 确保 Sync 被调用
	defer func() {
		_ = logger.Sync()
	}()

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
	newLogger := l.Logger.With(zap.Error(err))
	return &zapLogger{Logger: newLogger}
}

func (l *zapLogger) WithField(key string, value interface{}) Logger {
	newLogger := l.Logger.With(zap.Any(key, value))
	return &zapLogger{Logger: newLogger}
}

func (l *zapLogger) WithFields(fields map[string]interface{}) Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	newLogger := l.Logger.With(zapFields...)
	return &zapLogger{Logger: newLogger}
}

func (l *zapLogger) Log(ctx context.Context, level Level, args ...interface{}) {
	l.withContext(ctx).Log(levelToZapLevel(level), fmt.Sprint(args...))
}

func (l *zapLogger) Trace(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Debug(fmt.Sprint(args...))
}

func (l *zapLogger) Debug(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Debug(fmt.Sprint(args...))
}

func (l *zapLogger) Info(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Info(fmt.Sprint(args...))
}

func (l *zapLogger) Warn(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Warn(fmt.Sprint(args...))
}

func (l *zapLogger) Warning(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Warn(fmt.Sprint(args...))
}

func (l *zapLogger) Print(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Info(fmt.Sprint(args...))
}

func (l *zapLogger) Error(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Error(fmt.Sprint(args...))
}

func (l *zapLogger) Panic(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Panic(fmt.Sprint(args...))
}

func (l *zapLogger) Fatal(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Fatal(fmt.Sprint(args...))
}

func (l *zapLogger) Logf(ctx context.Context, level Level, format string, args ...interface{}) {
	l.withContext(ctx).Log(levelToZapLevel(level), fmt.Sprintf(format, args...))
}

func (l *zapLogger) Tracef(ctx context.Context, format string, args ...interface{}) {
	l.withContext(ctx).Debug(fmt.Sprintf(format, args...))
}

func (l *zapLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.withContext(ctx).Debug(fmt.Sprintf(format, args...))
}

func (l *zapLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.withContext(ctx).Info(fmt.Sprintf(format, args...))
}

func (l *zapLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.withContext(ctx).Warn(fmt.Sprintf(format, args...))
}

func (l *zapLogger) Warningf(ctx context.Context, format string, args ...interface{}) {
	l.withContext(ctx).Warn(fmt.Sprintf(format, args...))
}

func (l *zapLogger) Printf(ctx context.Context, format string, args ...interface{}) {
	l.withContext(ctx).Info(fmt.Sprintf(format, args...))
}

func (l *zapLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.withContext(ctx).Error(fmt.Sprintf(format, args...))
}

func (l *zapLogger) Panicf(ctx context.Context, format string, args ...interface{}) {
	l.withContext(ctx).Panic(fmt.Sprintf(format, args...))
}

func (l *zapLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	l.withContext(ctx).Fatal(fmt.Sprintf(format, args...))
}

func (l *zapLogger) Logln(ctx context.Context, level Level, args ...interface{}) {
	l.withContext(ctx).Log(levelToZapLevel(level), fmt.Sprintln(args...))
}

func (l *zapLogger) Traceln(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Debug(fmt.Sprintln(args...))
}

func (l *zapLogger) Debugln(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Debug(fmt.Sprintln(args...))
}

func (l *zapLogger) Infoln(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Info(fmt.Sprintln(args...))
}

func (l *zapLogger) Warnln(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Warn(fmt.Sprintln(args...))
}

func (l *zapLogger) Warningln(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Warn(fmt.Sprintln(args...))
}

func (l *zapLogger) Println(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Info(fmt.Sprintln(args...))
}

func (l *zapLogger) Errorln(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Error(fmt.Sprintln(args...))
}

func (l *zapLogger) Panicln(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Panic(fmt.Sprintln(args...))
}

func (l *zapLogger) Fatalln(ctx context.Context, args ...interface{}) {
	l.withContext(ctx).Fatal(fmt.Sprintln(args...))
}
