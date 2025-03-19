package logger

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pkg/errors"
)

type slogHook struct {
	handler                                 slog.Handler
	handlerOpts                             *slog.HandlerOptions
	opt                                     Options
	rotateNormalHandler, rotateErrHandler   slog.Handler
	localFsNormalHandler, localFsErrHandler slog.Handler
}

var _ slog.Handler = &slogHook{}

func newSlogHook(handler slog.Handler, handlerOpts *slog.HandlerOptions, opt Options) *slogHook {

	hook := &slogHook{
		handler:     handler,
		handlerOpts: handlerOpts,
		opt:         opt,
	}

	if opt.LocalFsConfig.Path != "" {
		normalWriter, errWriter, err := newLocalFsWriter(opt.LocalFsConfig)
		if err != nil {
			panic(fmt.Sprintf("new slog local fs writer error: %s", err))
		}

		hook.localFsNormalHandler = slog.NewJSONHandler(normalWriter, handlerOpts)
		hook.localFsErrHandler = slog.NewJSONHandler(errWriter, handlerOpts)
	}

	if opt.RotateConfig.Path != "" {
		normalWriter, errWriter, err := newRotateWriter(opt.RotateConfig)
		if err != nil {
			panic(fmt.Sprintf("new slog rotate writer error: %s", err))
		}

		hook.rotateNormalHandler = slog.NewJSONHandler(normalWriter, handlerOpts)
		hook.rotateErrHandler = slog.NewJSONHandler(errWriter, handlerOpts)
	}

	return hook
}

func (h *slogHook) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := h.handler.WithAttrs(attrs)

	newHook := &slogHook{
		handler:     newHandler,
		handlerOpts: h.handlerOpts,
		opt:         h.opt,
	}

	if h.rotateNormalHandler != nil {
		newHook.rotateNormalHandler = h.rotateNormalHandler.WithAttrs(attrs)
	}

	if h.rotateErrHandler != nil {
		newHook.rotateErrHandler = h.rotateErrHandler.WithAttrs(attrs)
	}

	if h.localFsNormalHandler != nil {
		newHook.localFsNormalHandler = h.localFsNormalHandler.WithAttrs(attrs)
	}

	if h.localFsErrHandler != nil {
		newHook.localFsErrHandler = h.localFsErrHandler.WithAttrs(attrs)
	}

	return newHook
}

func (h *slogHook) WithGroup(name string) slog.Handler {
	newHandler := h.handler.WithGroup(name)

	newHook := &slogHook{
		handler:     newHandler,
		handlerOpts: h.handlerOpts,
		opt:         h.opt,
	}

	if h.rotateNormalHandler != nil {
		newHook.rotateNormalHandler = h.rotateNormalHandler.WithGroup(name)
	}

	if h.rotateErrHandler != nil {
		newHook.rotateErrHandler = h.rotateErrHandler.WithGroup(name)
	}

	if h.localFsNormalHandler != nil {
		newHook.localFsNormalHandler = h.localFsNormalHandler.WithGroup(name)
	}

	if h.localFsErrHandler != nil {
		newHook.localFsErrHandler = h.localFsErrHandler.WithGroup(name)
	}

	return newHook
}

func (h *slogHook) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *slogHook) Handle(ctx context.Context, r slog.Record) error {
	if h.opt.Caller {
		caller := getCaller(h.opt.CallerSkip + 5)
		r.AddAttrs(slog.String(FieldKeyCaller, caller))
	}

	if h.opt.Timestamp {
		r.AddAttrs(slog.Int64(FieldKeyTimestamp, r.Time.Unix()))
	}

	if h.opt.ServiceName != "" {
		r.AddAttrs(slog.String(FieldKeyServiceName, h.opt.ServiceName))
	}

	if h.opt.RotateConfig.Path != "" {
		var handler slog.Handler
		if r.Level >= levelToSlogLevel(h.opt.RotateConfig.ErrorFileLevel) {
			handler = h.rotateErrHandler
		} else {
			handler = h.rotateNormalHandler
		}

		if handler != nil {
			err := handler.Handle(ctx, r)
			if err != nil {
				return errors.Wrapf(err, "slog rotate handle error")
			}
		}
	}

	if h.opt.LocalFsConfig.Path != "" {
		var handler slog.Handler
		if r.Level >= levelToSlogLevel(h.opt.LocalFsConfig.ErrorFileLevel) {
			handler = h.localFsErrHandler
		} else {
			handler = h.localFsNormalHandler
		}

		if handler != nil {
			err := handler.Handle(ctx, r)
			if err != nil {
				return errors.Wrapf(err, "slog local fs handle error")
			}
		}
	}

	return h.handler.Handle(ctx, r)
}
