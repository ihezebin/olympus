package logger

import (
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap/zapcore"
)

type logrusServiceHook struct {
	ServiceName string
}

var _ logrus.Hook = &logrusServiceHook{}

func newLogrusServiceHook(serviceName string) *logrusServiceHook {
	return &logrusServiceHook{
		ServiceName: serviceName,
	}
}

// Levels implement levels
func (hook *logrusServiceHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire implement fire
func (hook *logrusServiceHook) Fire(entry *logrus.Entry) error {
	entry.Data[FieldKeyServiceName] = hook.ServiceName
	return nil
}

type zerologServiceHook struct {
	ServiceName string
}

var _ zerolog.Hook = &zerologServiceHook{}

func newZerologServiceHook(serviceName string) *zerologServiceHook {
	return &zerologServiceHook{
		ServiceName: serviceName,
	}
}

func (hook *zerologServiceHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Str(FieldKeyServiceName, hook.ServiceName)
}

type zapServiceHook struct {
	ServiceName string
	core        zapcore.Core
}

var _ zapcore.Core = &zapServiceHook{}

func newZapServiceHook(core zapcore.Core, serviceName string) *zapServiceHook {
	return &zapServiceHook{
		ServiceName: serviceName,
		core:        core,
	}
}

func (h *zapServiceHook) With(fields []zapcore.Field) zapcore.Core {
	return h
}

func (h *zapServiceHook) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(entry, h)
}

func (h *zapServiceHook) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	newFields := append(fields, zapcore.Field{
		Key:    FieldKeyServiceName,
		Type:   zapcore.StringType,
		String: h.ServiceName,
	})
	if h.core != nil {
		return h.core.Write(ent, newFields)
	}
	return nil
}

func (h *zapServiceHook) Enabled(level zapcore.Level) bool {
	return true
}

func (h *zapServiceHook) Sync() error {
	if h.core != nil {
		return h.core.Sync()
	}
	return nil
}
