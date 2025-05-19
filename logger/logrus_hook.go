package logger

import (
	"context"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
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

type logrusCallerHook struct{}

var _ logrus.Hook = &logrusCallerHook{}

func newLogrusCallerHook() *logrusCallerHook {
	return &logrusCallerHook{}
}

func (h *logrusCallerHook) Fire(entry *logrus.Entry) error {
	entry.Data[FieldKeyCaller] = getCaller()
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

type logrusOtlpHook struct{}

var _ logrus.Hook = &logrusOtlpHook{}

func newLogrusOtlpHook() *logrusOtlpHook {
	return &logrusOtlpHook{}
}

func (h *logrusOtlpHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *logrusOtlpHook) Fire(entry *logrus.Entry) error {
	otelogger := global.Logger("logrus")
	record := log.Record{}

	attrs := make([]log.KeyValue, 0, len(entry.Data))
	for k, v := range entry.Data {
		attrs = append(attrs, log.String(k, fmt.Sprint(v)))
	}
	record.AddAttributes(attrs...)
	record.SetTimestamp(entry.Time)
	record.SetSeverity(h.convertLevel(entry.Level))
	record.SetSeverityText(entry.Level.String())
	record.SetEventName(entry.Level.String())
	record.SetBody(log.StringValue(entry.Message))

	// Collector 默认路径 "/v1/logs"，格式：
	// collectLogs "go.opentelemetry.io/proto/otlp/collector/logs/v1" collectLogs.ExportLogsServiceRequest
	otelogger.Emit(entry.Context, record)
	return nil
}

func (h *logrusOtlpHook) convertLevel(level logrus.Level) log.Severity {
	switch level {
	case logrus.PanicLevel:
		return log.SeverityFatal
	case logrus.FatalLevel:
		return log.SeverityFatal
	case logrus.ErrorLevel:
		return log.SeverityError
	case logrus.WarnLevel:
		return log.SeverityWarn
	case logrus.InfoLevel:
		return log.SeverityInfo
	case logrus.DebugLevel:
		return log.SeverityDebug
	case logrus.TraceLevel:
		return log.SeverityTrace
	default:
		return log.SeverityUndefined
	}
}
