package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ihezebin/rotatelog"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type logrusLocalFsHook struct {
	normalWriter, errWriter *os.File
	errLevel                logrus.Level
}

var _ logrus.Hook = &logrusLocalFsHook{}

func newLogrusLocalFsHook(config LocalFsConfig) (*logrusLocalFsHook, error) {
	path := config.Path
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, errors.Wrapf(err, "make dir:%s error", dir)
	}

	normalExt := filepath.Ext(path)
	errExt := fmt.Sprintf(".%s%s", config.ErrorFileExt, normalExt)
	errExt = strings.ReplaceAll(errExt, "..", ".")
	errPath := strings.ReplaceAll(path, normalExt, errExt)

	normalWriter, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, errors.Wrapf(err, "open normal file:%s error", path)
	}
	errWriter, err := os.OpenFile(errPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, errors.Wrapf(err, "open error file:%s error", errPath)
	}

	return &logrusLocalFsHook{
		normalWriter: normalWriter,
		errWriter:    errWriter,
		errLevel:     levelToLogrusLevel(config.ErrorFileLevel),
	}, nil
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

func newRotateWriter(config RotateConfig) (io.Writer, io.Writer, error) {
	path := config.Path
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, nil, errors.Wrapf(err, "make dir:%s error", dir)
	}

	normalExt := filepath.Ext(path)
	errExt := fmt.Sprintf(".%s%s", config.ErrorFileExt, normalExt)
	errExt = strings.ReplaceAll(errExt, "..", ".")
	errPath := strings.ReplaceAll(path, normalExt, errExt)

	rotatelog.BackupFilenameSeparator = "."

	normalWriter := &rotatelog.Rotater{
		Filename:   path,
		MaxSize:    config.MaxSizeKB,
		MaxBackups: config.MaxRetainFileCount,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
		LocalTime:  true,
	}
	errWriter := &rotatelog.Rotater{
		Filename:   errPath,
		MaxSize:    config.MaxSizeKB,
		MaxBackups: config.MaxRetainFileCount,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
		LocalTime:  true,
	}

	return normalWriter, errWriter, nil
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
