package logger

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
)

type slogHook struct {
	handler                                 slog.Handler
	handlerOpts                             *slog.HandlerOptions
	opt                                     Options
	rotateNormalHandler, rotateErrHandler   slog.Handler
	localFsNormalHandler, localFsErrHandler slog.Handler
	otlpAttrs                               []slog.Attr
}

var _ slog.Handler = &slogHook{}

func newSlogHook(handler slog.Handler, handlerOpts *slog.HandlerOptions, opt Options) *slogHook {

	hook := &slogHook{
		handler:     handler,
		handlerOpts: handlerOpts,
		opt:         opt,
		otlpAttrs:   make([]slog.Attr, 0),
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
		otlpAttrs:   append(h.otlpAttrs, attrs...),
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
		otlpAttrs:   h.otlpAttrs,
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
		caller := getCaller()
		r.AddAttrs(slog.String(FieldKeyCaller, caller))
	}

	if h.opt.Timestamp {
		r.AddAttrs(slog.Int64(FieldKeyTimestamp, r.Time.Unix()))
	}

	if h.opt.ServiceName != "" {
		r.AddAttrs(slog.String(FieldKeyServiceName, h.opt.ServiceName))
	}

	if h.opt.GetTraceIdFunc != nil {
		traceId := h.opt.GetTraceIdFunc(ctx)
		if traceId != "" {
			r.AddAttrs(slog.String(FieldKeyTraceId, traceId))
		}
	}

	if h.rotateNormalHandler != nil || h.rotateErrHandler != nil {
		var rotateHandler slog.Handler
		if r.Level >= levelToSlogLevel(h.opt.RotateConfig.ErrorFileLevel) {
			rotateHandler = h.rotateErrHandler
		} else {
			rotateHandler = h.rotateNormalHandler
		}

		if rotateHandler != nil {
			err := rotateHandler.Handle(ctx, r)
			if err != nil {
				return errors.Wrapf(err, "slog rotate handle error")
			}
		}
	}

	if h.localFsNormalHandler != nil || h.localFsErrHandler != nil {
		var localFsHandler slog.Handler
		if r.Level >= levelToSlogLevel(h.opt.LocalFsConfig.ErrorFileLevel) {
			localFsHandler = h.localFsErrHandler
		} else {
			localFsHandler = h.localFsNormalHandler
		}

		if localFsHandler != nil {
			err := localFsHandler.Handle(ctx, r)
			if err != nil {
				return errors.Wrapf(err, "slog local fs handle error")
			}
		}
	}

	if h.opt.OtlpEnabled {
		otelogger := global.Logger("slog")
		record := log.Record{}

		attrs := make([]log.KeyValue, 0, len(h.otlpAttrs))
		for _, attr := range h.otlpAttrs {
			attrs = append(attrs, log.String(attr.Key, fmt.Sprint(attr.Value.String())))
		}
		record.AddAttributes(attrs...)
		record.SetTimestamp(r.Time)
		record.SetSeverity(h.convertLevel2OtlpLevel(r.Level))
		record.SetSeverityText(r.Level.String())
		record.SetEventName(r.Level.String())
		record.SetBody(log.StringValue(r.Message))

		// Collector 默认路径 "/v1/logs"，格式：
		// collectLogs "go.opentelemetry.io/proto/otlp/collector/logs/v1" collectLogs.ExportLogsServiceRequest
		otelogger.Emit(ctx, record)
	}

	return h.handler.Handle(ctx, r)
}

func (h *slogHook) convertLevel2OtlpLevel(level slog.Level) log.Severity {
	switch level {
	case slog.LevelDebug:
		return log.SeverityDebug
	case slog.LevelInfo:
		return log.SeverityInfo
	case slog.LevelWarn:
		return log.SeverityWarn
	case slog.LevelError:
		return log.SeverityError
	default:
		return log.SeverityUndefined
	}
}
