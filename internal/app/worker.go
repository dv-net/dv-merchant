package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dv-net/dv-merchant/internal/metrics"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

const (
	workerRestartBaseDelay = 1 * time.Second
	workerRestartMaxDelay  = 30 * time.Second
)

type worker struct {
	name string
	run  func(context.Context) error
}

func superviseWorkers(ctx context.Context, lg logger.Logger, workers ...worker) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	for _, w := range workers {
		wg.Add(1)
		go func(w worker) {
			defer wg.Done()
			superviseWorker(ctx, lg, w)
		}(w)
	}
	return wg
}

func superviseWorker(ctx context.Context, lg logger.Logger, w worker) {
	delay := workerRestartBaseDelay
	for {
		if ctx.Err() != nil {
			return
		}

		start := time.Now()
		err := runWorkerOnce(ctx, w)
		switch {
		case ctx.Err() != nil:
			// Shutdown in progress - a returning worker is expected.
			return
		case err != nil:
			metrics.IncBackgroundWorkerRestarts(w.name)
			lg.Errorw("background worker crashed, restarting", "worker", w.name, "uptime", time.Since(start).String(), "error", err)
		default:
			metrics.IncBackgroundWorkerRestarts(w.name)
			lg.Warnw("background worker exited unexpectedly, restarting", "worker", w.name, "uptime", time.Since(start).String())
		}

		if time.Since(start) >= workerRestartMaxDelay {
			delay = workerRestartBaseDelay
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}

		if delay *= 2; delay > workerRestartMaxDelay {
			delay = workerRestartMaxDelay
		}
	}
}

func runWorkerOnce(ctx context.Context, w worker) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	metrics.SetBackgroundWorkerRunning(w.name, true)
	defer metrics.SetBackgroundWorkerRunning(w.name, false)

	return w.run(ctx)
}

func asWorker(name string, run func(context.Context)) worker {
	return worker{name: name, run: func(ctx context.Context) error {
		run(ctx)
		return nil
	}}
}
