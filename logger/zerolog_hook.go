package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
)

type zerologRotateHook struct {
	normalLogger, errLogger zerolog.Logger
	errLevel                zerolog.Level
}

var _ zerolog.Hook = &zerologRotateHook{}

func newZerologRotateHook(logger zerolog.Logger, config RotateConfig) (*zerologRotateHook, error) {
	normalWriter, errWriter, err := newRotateWriter(config)
	if err != nil {
		return nil, errors.Wrapf(err, "new writer error")
	}

	// 使用 With() 创建一个新的 logger
	normalLogger := logger.Output(normalWriter)
	errLogger := logger.Output(errWriter)

	return &zerologRotateHook{
		normalLogger: normalLogger,
		errLogger:    errLogger,
		errLevel:     levelToZerologLevel(config.ErrorFileLevel),
	}, nil
}

func (h *zerologRotateHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	var logger zerolog.Logger
	if level >= h.errLevel {
		logger = h.errLogger
	} else {
		logger = h.normalLogger
	}

	logger.WithLevel(level).Msg(msg)
}

type zerologLocalFsHook struct {
	normalLogger, errLogger zerolog.Logger
	errLevel                zerolog.Level
}

var _ zerolog.Hook = &zerologLocalFsHook{}

func newZerologLocalFsHook(logger zerolog.Logger, config LocalFsConfig) *zerologLocalFsHook {
	normalWriter, errWriter, err := newLocalFsWriter(config)
	if err != nil {
		panic(fmt.Sprintf("new local fs writer error: %s", err))
	}

	normalLogger := logger.Output(normalWriter)
	errLogger := logger.Output(errWriter)

	return &zerologLocalFsHook{
		normalLogger: normalLogger,
		errLogger:    errLogger,
		errLevel:     levelToZerologLevel(config.ErrorFileLevel),
	}
}

func (h *zerologLocalFsHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	var logger zerolog.Logger
	if level >= h.errLevel {
		logger = h.errLogger
	} else {
		logger = h.normalLogger
	}

	logger.WithLevel(level).Msg(msg)
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

type zerologTimestampHook struct{}

var _ zerolog.Hook = &zerologTimestampHook{}

func newZerologTimestampHook() zerolog.Hook {
	return &zerologTimestampHook{}
}

func (t zerologTimestampHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Int64(FieldKeyTimestamp, time.Now().Unix())
}

type zerologTraceIdHook struct {
	GetTraceIdFunc func(ctx context.Context) string
}

var _ zerolog.Hook = &zerologTraceIdHook{}

func newZerologTraceIdHook(getTraceIdFunc func(ctx context.Context) string) *zerologTraceIdHook {
	return &zerologTraceIdHook{
		GetTraceIdFunc: getTraceIdFunc,
	}
}

func (h *zerologTraceIdHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	traceId := h.GetTraceIdFunc(ctx)
	if traceId != "" {
		e.Str(FieldKeyTraceId, traceId)
	}
}

type zerologCallerHook struct {
}

var _ zerolog.Hook = &zerologCallerHook{}

func newZerologCallerHook() zerolog.Hook {
	return &zerologCallerHook{}
}

func (h *zerologCallerHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Str(FieldKeyCaller, getCaller())
}

type zerologOtlpHook struct {
	fields map[string]interface{}
	err    error
}

var _ zerolog.Hook = &zerologOtlpHook{}

func newZerologOtlpHook(fields map[string]interface{}, err error) zerolog.Hook {
	return &zerologOtlpHook{
		fields: fields,
		err:    err,
	}
}

func (h *zerologOtlpHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	otelogger := global.Logger("zerolog")
	record := log.Record{}
	attrs := make([]log.KeyValue, 0, len(h.fields))
	for k, v := range h.fields {
		attrs = append(attrs, log.String(k, fmt.Sprint(v)))
	}
	record.AddAttributes(attrs...)
	record.SetTimestamp(zerolog.TimestampFunc())
	record.SetSeverity(h.convertLevel2OtlpLevel(level))
	record.SetSeverityText(level.String())
	record.SetEventName(msg)
	record.SetBody(log.StringValue(msg))

	otelogger.Emit(e.GetCtx(), record)
}

func (h *zerologOtlpHook) convertLevel2OtlpLevel(level zerolog.Level) log.Severity {
	switch level {
	case zerolog.DebugLevel:
		return log.SeverityDebug
	case zerolog.InfoLevel:
		return log.SeverityInfo
	case zerolog.WarnLevel:
		return log.SeverityWarn
	case zerolog.ErrorLevel:
		return log.SeverityError
	case zerolog.FatalLevel:
		return log.SeverityFatal
	case zerolog.PanicLevel:
		return log.SeverityFatal
	default:
		return log.SeverityUndefined
	}
}
