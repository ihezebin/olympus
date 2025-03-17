package logger

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

var ctx = context.Background()

func TestLogger(t *testing.T) {
	ResetLoggerWithOptions(WithPrettyCallerHook(), WithTimestampHook(), WithServiceName("unit_test"))
	WithField("key", "value").Info(ctx, "hello")

	logger := logrus.New()
	logger.SetReportCaller(true)
	logger.Info("hello")
}

func TestLocalFsHook(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	filename := filepath.Join(pwd, "localfs.log")

	t.Log(pwd, filename)
	ResetLoggerWithOptions(WithCallerHook(), WithLocalFsHook(filename))

	Info(ctx, "hello world")
	Errorf(ctx, "hello world error")
}

func TestRotateLogsFsHook(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	filename := filepath.Join(pwd, "rotatelogs.log")

	t.Log(pwd, filename)

	ResetLoggerWithOptions(WithCallerHook(), WithRotateLogsHook(filename, time.Second*3, time.Second*9))

	for i := 0; i < 12; i++ {
		Info(ctx, "hello world ", time.Now().Format(time.DateTime))
		Error(ctx, "hello world error ", time.Now().Format(time.DateTime))
		time.Sleep(time.Second)
	}
}
