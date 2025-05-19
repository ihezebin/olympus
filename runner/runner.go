package runner

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ihezebin/olympus/logger"
)

type Runner struct {
	tasks []Task
}

// Task 必须为阻塞型任务，持续运行，否则会中断其他任务
type Task interface {
	Name() string
	Run(ctx context.Context) (err error)
	Close(ctx context.Context) (err error)
}

func NewRunner(tasks ...Task) *Runner {
	return &Runner{tasks: tasks}
}

// Run 阻塞运行，直到收到 SIGTERM, SIGQUIT, SIGINT 信号
func (r *Runner) Run(ctx context.Context) {
	r.run(ctx, false)
}

// RunWithOneStop 阻塞运行，直到收到 SIGTERM, SIGQUIT, SIGINT 信号，或者其中一个任务意外停止
func (r *Runner) RunWithOneStop(ctx context.Context) {
	r.run(ctx, true)
}

func (r *Runner) run(ctx context.Context, oneStop bool) {
	ctx, cancel := context.WithCancel(ctx)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	defer func() {
		signal.Stop(ch)
		close(ch)
	}()
	go func() {
		logger.Infof(ctx, "got signal %v, will cancel all tasks", <-ch)
		cancel()
		for _, t := range r.tasks {
			t.Close(ctx)
		}
		logger.Info(ctx, "all tasks closed")
	}()

	var wg sync.WaitGroup
	for _, t := range r.tasks {
		wg.Add(1)
		go func(task Task) {
			defer wg.Done()
			logger.Infof(ctx, "task(%s) is starting", task.Name())
			if err := task.Run(ctx); err != nil {
				logger.Errorf(ctx, "task(%s) run with error(%v)", task.Name(), err)
				// one task stop, cancel the context to stop all other tasks.
				if oneStop {
					logger.Infof(ctx, "task(%s) err, will cancel all tasks", task.Name())
					cancel()
				}
			}
			logger.Infof(ctx, "task(%s) is stopped", task.Name())
		}(t)
	}
	wg.Wait()
}
