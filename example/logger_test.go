package example

import (
	"context"
	"testing"

	"github.com/ihezebin/olympus/logger"
)

func TestLogger(t *testing.T) {
	ctx := context.Background()

	logger.Error(ctx, "hello")
	logger.WithField("key", "value").Info(ctx, "hello")
}
