package example

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ihezebin/olympus/logger"
)

func TestLogger(t *testing.T) {

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	logger.ResetLoggerWithOptions(
		logger.WithLoggerType(logger.LoggerTypeLogrus),
		logger.WithServiceName("example"),
		logger.WithCaller(),
		logger.WithCallerSkip(1),
		logger.WithTimestamp(),
		logger.WithLevel(logger.LevelDebug),
		//logger.WithLocalFsHook(filepath.Join(conf.Pwd, conf.Logger.Filename)),
		// 每天切割，保留 3 天的日志
		logger.WithRotate(logger.RotateConfig{
			Path:               filepath.Join(pwd, "logs/example.log"),
			MaxSizeKB:          1024 * 500, // 500 MB
			MaxAge:             time.Hour * 24 * 7,
			MaxRetainFileCount: 3,
			Compress:           true,
		}),
	)

	ctx := context.Background()

	logger.Error(ctx, "hello")
	logger.WithField("key", "value").Info(ctx, "hello")
}
