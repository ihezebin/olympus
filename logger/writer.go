package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ihezebin/rotatelog"
	"github.com/pkg/errors"
)

func newLocalFsWriter(config LocalFsConfig) (io.Writer, io.Writer, error) {
	path := config.Path
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, nil, errors.Wrapf(err, "make dir:%s error", dir)
	}

	normalExt := filepath.Ext(path)
	errExt := fmt.Sprintf(".%s%s", config.ErrorFileExt, normalExt)
	errExt = strings.ReplaceAll(errExt, "..", ".")
	errPath := strings.ReplaceAll(path, normalExt, errExt)

	normalWriter, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "open normal file:%s error", path)
	}
	errWriter, err := os.OpenFile(errPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "open error file:%s error", errPath)
	}

	return normalWriter, errWriter, nil
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
