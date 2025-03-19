package logger

import (
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
	normalLogger := logger.With().CallerWithSkipFrameCount(opt.CallerSkip + 7).Logger().Output(normalWriter)

	errLogger := logger.With().CallerWithSkipFrameCount(opt.CallerSkip + 7).Logger().Output(errWriter)

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

	normalLogger := logger.With().CallerWithSkipFrameCount(opt.CallerSkip + 7).Logger().Output(normalWriter)
	errLogger := logger.With().CallerWithSkipFrameCount(opt.CallerSkip + 7).Logger().Output(errWriter)

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
