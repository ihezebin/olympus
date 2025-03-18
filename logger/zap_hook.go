package logger

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
)

type zapHook struct {
	core                                zapcore.Core
	opt                                 *Options
	rotateNormalWriter, rotateErrWriter io.Writer
	encoder                             zapcore.Encoder
}

var _ zapcore.Core = &zapHook{}

func newZapHook(core zapcore.Core, encoder zapcore.Encoder, opt *Options) *zapHook {
	hook := &zapHook{core: core, encoder: encoder, opt: opt}
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
		caller := getCaller(h.opt.CallerSkip + 6)
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

	return h.core.Write(entry, newFields)
}

func (h *zapHook) Enabled(level zapcore.Level) bool {
	return true
}

func (h *zapHook) Sync() error {
	return h.core.Sync()
}
