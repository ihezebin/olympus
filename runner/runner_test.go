package runner

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/ihezebin/soup/httpserver"
)

/*
{"level":"info","msg":"task(httpserver) is starting","time":"2025-01-24 16:38:11"}
{"level":"info","msg":"http server is starting in port: 8080","time":"2025-01-24 16:38:11"}
{"level":"info","msg":"got signal interrupt, will cancel all tasks","time":"2025-01-24 16:38:15"}
{"level":"info","msg":"http server closed","time":"2025-01-24 16:38:15"}
{"level":"info","msg":"task(httpserver) is stopped","time":"2025-01-24 16:38:15"}
{"level":"info","msg":"all tasks closed","time":"2025-01-24 16:38:15"}
*/
func TestRunner(t *testing.T) {
	p := os.Getpid()
	process, err := os.FindProcess(p)
	if err != nil {
		t.Fatal("Failed to find process:", err)
	}

	go func() {
		time.Sleep(time.Second * 5)
		process.Signal(syscall.SIGINT)
	}()

	runner := NewRunner(httpserver.NewServer())
	runner.Run(context.Background())
}
