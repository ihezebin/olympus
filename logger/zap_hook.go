package logger

import (
	"context"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.uber.org/zap/zapcore"
)

type zapHook struct {
	ctx                                   context.Context
	core                                  zapcore.Core
	opt                                   Options
	rotateNormalWriter, rotateErrWriter   io.Writer
	localFsNormalWriter, localFsErrWriter io.Writer
	encoder                               zapcore.Encoder
}

var _ zapcore.Core = &zapHook{}

func newZapHook(ctx context.Context, core zapcore.Core, encoder zapcore.Encoder, opt Options) *zapHook {
	hook := &zapHook{ctx: ctx, core: core, encoder: encoder, opt: opt}

	if opt.LocalFsConfig.Path != "" {
		normalWriter, errWriter, err := newLocalFsWriter(opt.LocalFsConfig)
		if err != nil {
			panic(fmt.Sprintf("new zap local fs writer error: %s", err))
		}
		hook.localFsNormalWriter = normalWriter
		hook.localFsErrWriter = errWriter
	}

	if opt.RotateConfig.Path != "" {
		normalWriter, errWriter, err := newRotateWriter(opt.RotateConfig)
		if err != nil {
			panic(fmt.Sprintf("new zap rotate writer error: %s", err))
		}
		hook.rotateNormalWriter = normalWriter
		hook.rotateErrWriter = errWriter
	}
	return hook
}

func (h *zapHook) With(fields []zapcore.Field) zapcore.Core {
	newCore := h.core.With(fields)
	return &zapHook{
		ctx:                h.ctx,
		core:               newCore,
		opt:                h.opt,
		rotateNormalWriter: h.rotateNormalWriter,
		rotateErrWriter:    h.rotateErrWriter,
		encoder:            h.encoder,
	}
}

func (h *zapHook) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if h.Enabled(entry.Level) {
		ce = ce.AddCore(entry, h)
	}
	return ce
}

func (h *zapHook) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	newFields := make([]zapcore.Field, 0, len(fields))
	newFields = append(newFields, fields...)

	if h.opt.Caller {
		caller := getCaller()
		newFields = append(newFields, zapcore.Field{
			Key:    FieldKeyCaller,
			Type:   zapcore.StringType,
			String: caller,
		})
	}

	if h.opt.Timestamp {
		newFields = append(newFields, zapcore.Field{
			Key:     FieldKeyTimestamp,
			Type:    zapcore.Int64Type,
			Integer: entry.Time.Unix(),
		})
	}

	if h.opt.ServiceName != "" {
		newFields = append(newFields, zapcore.Field{
			Key:    FieldKeyServiceName,
			Type:   zapcore.StringType,
			String: h.opt.ServiceName,
		})
	}

	if h.opt.GetTraceIdFunc != nil {
		traceId := h.opt.GetTraceIdFunc(h.ctx)
		if traceId != "" {
			newFields = append(newFields, zapcore.Field{
				Key:    FieldKeyTraceId,
				Type:   zapcore.StringType,
				String: traceId,
			})
		}
	}

	if h.opt.RotateConfig.Path != "" {
		var writer io.Writer
		if entry.Level >= levelToZapLevel(h.opt.RotateConfig.ErrorFileLevel) {
			writer = h.rotateErrWriter
		} else {
			writer = h.rotateNormalWriter
		}

		buf, err := h.encoder.EncodeEntry(entry, newFields)
		if err != nil {
			return errors.Wrapf(err, "encode error")
		}
		defer buf.Free()

		_, err = writer.Write(buf.Bytes())
		if err != nil {
			return errors.Wrapf(err, "zap rotate write error")
		}
	}

	if h.localFsNormalWriter != nil {
		var writer io.Writer
		if entry.Level >= levelToZapLevel(h.opt.LocalFsConfig.ErrorFileLevel) {
			writer = h.localFsErrWriter
		} else {
			writer = h.localFsNormalWriter
		}

		buf, err := h.encoder.EncodeEntry(entry, newFields)
		if err != nil {
			return errors.Wrapf(err, "encode error")
		}
		defer buf.Free()

		_, err = writer.Write(buf.Bytes())
		if err != nil {
			return errors.Wrapf(err, "zap local fs write error")
		}
	}

	if h.opt.OtlpEnabled {
		otelogger := global.Logger("zap")
		record := log.Record{}

		attrs := make([]log.KeyValue, 0, len(newFields))
		for _, field := range newFields {
			attrs = append(attrs, log.String(field.Key, fmt.Sprint(field.String)))
		}
		record.AddAttributes(attrs...)
		record.SetTimestamp(entry.Time)
		record.SetSeverity(h.convertLevel2OtlpLevel(entry.Level))
		record.SetSeverityText(entry.Level.String())
		record.SetEventName(entry.Level.String())
		record.SetBody(log.StringValue(entry.Message))

		ctx := h.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		otelogger.Emit(ctx, record)
	}

	return h.core.Write(entry, newFields)
}

func (h *zapHook) Enabled(level zapcore.Level) bool {
	return h.core.Enabled(level)
}

func (h *zapHook) Sync() error {
	return h.core.Sync()
}

func (h *zapHook) convertLevel2OtlpLevel(level zapcore.Level) log.Severity {
	switch level {
	case zapcore.DebugLevel:
		return log.SeverityDebug
	case zapcore.InfoLevel:
		return log.SeverityInfo
	case zapcore.WarnLevel:
		return log.SeverityWarn
	case zapcore.ErrorLevel:
		return log.SeverityError
	case zapcore.FatalLevel:
		return log.SeverityFatal
	case zapcore.PanicLevel:
		return log.SeverityFatal
	default:
		return log.SeverityUndefined
	}
}
