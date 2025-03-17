package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
