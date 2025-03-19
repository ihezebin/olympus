package example

import (
	"context"
	"testing"

	"github.com/ihezebin/olympus/logger"
)

func TestLogger(t *testing.T) {
	ctx := context.Background()

	logger.Info(ctx, "hello")
}
