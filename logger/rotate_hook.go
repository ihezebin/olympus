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

var _ logrus.Hook = &logrusRotateHook{}

type logrusRotateHook struct {
	normalWriter, errWriter io.Writer
	errLevel                logrus.Level
}

func newLogrusRotateHook(config RotateConfig) (*logrusRotateHook, error) {
	normalWriter, errWriter, err := newWriter(config)
	if err != nil {
		return nil, errors.Wrapf(err, "new writer error")
	}

	return &logrusRotateHook{
		normalWriter: normalWriter,
		errWriter:    errWriter,
		errLevel:     levelToLogrusLevel(config.ErrorFileLevel),
	}, nil
}

func newWriter(config RotateConfig) (io.Writer, io.Writer, error) {
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
