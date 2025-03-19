package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type zerologRotateHook struct {
	normalLogger, errLogger zerolog.Logger
	errLevel                zerolog.Level
}

var _ zerolog.Hook = &zerologRotateHook{}

func newZerologRotateHook(logger zerolog.Logger, opt Options, config RotateConfig) (*zerologRotateHook, error) {
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

func newZerologLocalFsHook(logger zerolog.Logger, opt Options, config LocalFsConfig) *zerologLocalFsHook {
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
