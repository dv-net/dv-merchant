package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	ctx, cancel := signal.NotifyContext(
		ctx,
		[]os.Signal{
			syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT,
		}...,
	)

	defer func() {
		cancel()
	}()

	st, err := storage.InitStore(ctx, conf)
	if err != nil {
		lg.Error("failed to init store", err)
		return err
	}
	defer func() {
		if storageCloseErr := st.Close(); storageCloseErr != nil {
			lg.Error("storage close error", storageCloseErr)
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
				lg.Error("failed to start service", err)
			}
		}()
	}

	srv := server.NewServer(conf.HTTP, services, lg)

	lg.Info("Dv-merchant Server Start")

	if err := srv.Stop(); err != nil {
		lg.Error("failed to stop server", err)
	}

	serverErrCh := make(chan error, 1)
	go func() {
		defer close(serverErrCh)
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case srvErr := <-serverErrCh:
		return srvErr
	}
}
