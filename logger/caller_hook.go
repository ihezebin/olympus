package logger

import (
	"fmt"
	"runtime"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap/zapcore"
)

type logrusCallerHook struct {
	skipFrameCount int
}

func newLogrusCallerHook(skipFrameCount int) *logrusCallerHook {
	return &logrusCallerHook{skipFrameCount: skipFrameCount}
}

func (h *logrusCallerHook) Fire(entry *logrus.Entry) error {
	entry.Data[FieldKeyCaller] = getCaller(h.skipFrameCount)
	return nil
}

func (h *logrusCallerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func getCaller(skipFrameCount int) string {
	pc := make([]uintptr, 1)
	n := runtime.Callers(skipFrameCount+1, pc)
	if n == 0 {
		return ""
	}
	frames := runtime.CallersFrames(pc)
	frame, _ := frames.Next()
	return fmt.Sprintf("%s:%d", frame.File, frame.Line)
}

type zapCallerHook struct {
	skipFrameCount int
	core           zapcore.Core
}

var _ zapcore.Core = &zapCallerHook{}

func newZapCallerHook(core zapcore.Core, skipFrameCount int) *zapCallerHook {
	return &zapCallerHook{core: core, skipFrameCount: skipFrameCount}
}

func (h *zapCallerHook) With(fields []zapcore.Field) zapcore.Core {
	return h
}

func (h *zapCallerHook) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(entry, h)
}

func (h *zapCallerHook) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	caller := getCaller(h.skipFrameCount)
	newFields := append(fields, zapcore.Field{
		Key:    FieldKeyCaller,
		Type:   zapcore.StringType,
		String: caller,
	})
	if h.core != nil {
		return h.core.Write(ent, newFields)
	}
	return nil
}

func (h *zapCallerHook) Enabled(level zapcore.Level) bool {
	return true
}

func (h *zapCallerHook) Sync() error {
	if h.core != nil {
		return h.core.Sync()
	}
	return nil
}
