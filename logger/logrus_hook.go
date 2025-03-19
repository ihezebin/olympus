package logger

import (
	"context"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type logrusLocalFsHook struct {
	normalWriter, errWriter io.Writer
	errLevel                logrus.Level
}

var _ logrus.Hook = &logrusLocalFsHook{}

func newLogrusLocalFsHook(config LocalFsConfig) *logrusLocalFsHook {
	normalWriter, errWriter, err := newLocalFsWriter(config)
	if err != nil {
		panic(fmt.Sprintf("new local fs writer error: %s", err))
	}

	return &logrusLocalFsHook{
		normalWriter: normalWriter,
		errWriter:    errWriter,
		errLevel:     levelToLogrusLevel(config.ErrorFileLevel),
	}
}

func (l *logrusLocalFsHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (l *logrusLocalFsHook) Fire(entry *logrus.Entry) error {
	data, err := entry.Logger.Formatter.Format(entry)
	if err != nil {
		return errors.Wrapf(err, "format log error")
	}

	if entry.Level <= l.errLevel {
		_, err = l.errWriter.Write(data)
		if err != nil {
			return errors.Wrapf(err, "write error log error")
		}
	} else {
		_, err = l.normalWriter.Write(data)
		if err != nil {
			return errors.Wrapf(err, "write normal log error")
		}
	}
	return nil
}

type logrusCallerHook struct {
	skipFrameCount int
}

var _ logrus.Hook = &logrusCallerHook{}

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

type logrusServiceHook struct {
	ServiceName string
}

var _ logrus.Hook = &logrusServiceHook{}

func newLogrusServiceHook(serviceName string) *logrusServiceHook {
	return &logrusServiceHook{
		ServiceName: serviceName,
	}
}

func (hook *logrusServiceHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *logrusServiceHook) Fire(entry *logrus.Entry) error {
	entry.Data[FieldKeyServiceName] = hook.ServiceName
	return nil
}

var _ logrus.Hook = &logrusRotateHook{}

type logrusRotateHook struct {
	normalWriter, errWriter io.Writer
	errLevel                logrus.Level
}

var _ logrus.Hook = &logrusRotateHook{}

func newLogrusRotateHook(config RotateConfig) (*logrusRotateHook, error) {
	normalWriter, errWriter, err := newRotateWriter(config)
	if err != nil {
		return nil, errors.Wrapf(err, "new writer error")
	}

	return &logrusRotateHook{
		normalWriter: normalWriter,
		errWriter:    errWriter,
		errLevel:     levelToLogrusLevel(config.ErrorFileLevel),
	}, nil
}

func (l *logrusRotateHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (l *logrusRotateHook) Fire(entry *logrus.Entry) error {
	data, err := entry.Logger.Formatter.Format(entry)
	if err != nil {
		return errors.Wrapf(err, "format log error")
	}

	if entry.Level <= l.errLevel {
		_, err = l.errWriter.Write(data)
		if err != nil {
			return errors.Wrapf(err, "write error log error")
		}
	} else {
		_, err = l.normalWriter.Write(data)
		if err != nil {
			return errors.Wrapf(err, "write normal log error")
		}
	}
	return nil
}

type logrusTraceIdHook struct {
	GetTraceIdFunc func(ctx context.Context) string
}

var _ logrus.Hook = &logrusTraceIdHook{}

func newLogrusTraceIdHook(getTraceIdFunc func(ctx context.Context) string) *logrusTraceIdHook {
	return &logrusTraceIdHook{
		GetTraceIdFunc: getTraceIdFunc,
	}
}

func (h *logrusTraceIdHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *logrusTraceIdHook) Fire(entry *logrus.Entry) error {
	traceId := h.GetTraceIdFunc(entry.Context)
	if traceId != "" {
		entry.Data[FieldKeyTraceId] = traceId
	}
	return nil
}
