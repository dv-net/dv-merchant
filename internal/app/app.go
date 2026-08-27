package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dv-net/dv-merchant/internal/cache"
	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/server"
	"github.com/dv-net/dv-merchant/internal/service"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/pkg/logger"
	logger2 "github.com/dv-net/mx/logger"
	"github.com/dv-net/mx/ops"
)

func Run(ctx context.Context, conf *config.Config, l logger.Logger, currentAppVersion, commitHash string) error {
	lg := l

	st, err := storage.InitStore(ctx, conf)
	if err != nil {
		lg.Errorw("failed to init store", "error", err)
		return err
	}
	defer func() {
		if storageCloseErr := st.Close(); storageCloseErr != nil {
			lg.Errorw("storage close error", "error", storageCloseErr)
		}
	}()

	dbSyncer := logger.NewDBWriteSyncer(st)
	lg.WithDBSyncer(dbSyncer)

	ca := cache.InitCache()

	services, err := service.NewServices(ctx, conf, st, ca, lg, currentAppVersion, commitHash)
	if err != nil {
		lg.Fatal("error start DI service", err)
	}

	workerCtx, stopWorkers := context.WithCancel(ctx)
	defer stopWorkers()
	workersWG := superviseWorkers(workerCtx, lg, buildWorkers(services, conf, st, lg)...)

	svcs := ops.New(logger2.NewExtended(), conf.Ops)

	for _, svc := range svcs {
		go func() {
			if err := svc.Start(ctx); err != nil {
				lg.Errorw("failed to start service", "error", err)
			}
		}()
	}

	srv := server.NewServer(conf.HTTP, services, lg)

	lg.Info("Dv-merchant Server Start")

	serverErrCh := make(chan error, 1)
	go func() {
		defer close(serverErrCh)
		if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		return shutdown(ctx, conf.HTTP.ShutdownTimeout, srv, stopWorkers, workersWG, lg)
	case err := <-serverErrCh:
		_ = shutdown(ctx, conf.HTTP.ShutdownTimeout, srv, stopWorkers, workersWG, lg)
		if err != nil {
			return fmt.Errorf("server: %w", err)
		}
		return nil
	}
}

func shutdown(ctx context.Context, timeout time.Duration, srv *server.Server, stopWorkers context.CancelFunc, workersWG *sync.WaitGroup, lg logger.Logger) error {
	if timeout <= 0 {
		timeout = 25 * time.Second
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()
	stopWorkers()

	var shutdownErr error
	if err := srv.Shutdown(shutdownCtx); err != nil {
		lg.Errorw("http server shutdown error", "error", err)
		shutdownErr = fmt.Errorf("server shutdown: %w", err)
	}

	done := make(chan struct{})
	go func() {
		workersWG.Wait()
		close(done)
	}()

	select {
	case <-done:
		lg.Info("background workers stopped")
	case <-shutdownCtx.Done():
		lg.Warnw("background workers did not stop before shutdown timeout", "timeout", timeout.String())
	}

	return shutdownErr
}
