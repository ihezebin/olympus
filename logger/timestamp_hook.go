package logger

import (
	"time"

	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap/zapcore"
)

type logrusTimestampHook struct{}

var _ logrus.Hook = &logrusTimestampHook{}

func newLogrusTimestampHook() logrus.Hook {
	return &logrusTimestampHook{}
}

func (t logrusTimestampHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (t logrusTimestampHook) Fire(entry *logrus.Entry) error {
	entry.Data[FieldKeyTimestamp] = entry.Time.Unix()
	return nil
}

type zerologTimestampHook struct{}

var _ zerolog.Hook = &zerologTimestampHook{}

func newZerologTimestampHook() zerolog.Hook {
	return &zerologTimestampHook{}
}

func (t zerologTimestampHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Int64(FieldKeyTimestamp, time.Now().Unix())
}

type zapTimestampHook struct {
	core zapcore.Core
}

func newZapTimestampHook(core zapcore.Core) zapcore.Core {
	return &zapTimestampHook{core: core}
}

func (h *zapTimestampHook) With(fields []zapcore.Field) zapcore.Core {
	return h.core.With(fields)
}

func (h *zapTimestampHook) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if h.Enabled(entry.Level) {
		return ce.AddCore(entry, h)
	}
	return ce
}

func (h *zapTimestampHook) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	newFields := append(fields, zapcore.Field{
		Key:     FieldKeyTimestamp,
		Type:    zapcore.Int64Type,
		Integer: ent.Time.Unix(),
	})
	if h.core != nil {
		return h.core.Write(ent, newFields)
	}
	return nil
}

func (h *zapTimestampHook) Enabled(level zapcore.Level) bool {
	return true
}

func (h *zapTimestampHook) Sync() error {
	if h.core != nil {
		return h.core.Sync()
	}
	return nil
}
