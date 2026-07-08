package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

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

	initTickers(ctx, services, conf, st, lg)

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

	if err := srv.Stop(); err != nil {
		lg.Errorw("failed to stop server", "error", err)
	}

	serverErrCh := make(chan error, 1)
	go func() {
		defer close(serverErrCh)
		if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		if err := srv.Stop(); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}
		return nil
	case err := <-serverErrCh:
		return fmt.Errorf("server: %w", err)
	}
}
