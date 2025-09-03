package updater

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

type IUpdater interface {
	CheckNewVersionBackend(ctx context.Context) (*VersionInfo, error)
	CheckNewVersionProcessing(ctx context.Context) (*VersionInfo, error)
	CheckApplicationVersions(ctx context.Context) (*ApplicationVersion, error)
	UpdateBackend(ctx context.Context) error
	UpdateProcessing(ctx context.Context) error
}

type Service struct {
	l                 logger.Logger
	cfg               *config.Config
	processing        processing.IProcessingSystem
	currentAppVersion string
	updaterClient     IUpdateClient
}

func New(l logger.Logger, cfg *config.Config, processing processing.IProcessingSystem, currentAppVersion string) *Service {
	updaterClient, _ := NewClient(l, cfg)
	return &Service{
		l:                 l,
		cfg:               cfg,
		processing:        processing,
		currentAppVersion: currentAppVersion,
		updaterClient:     updaterClient,
	}
}

func (s *Service) CheckNewVersionBackend(ctx context.Context) (*VersionInfo, error) {
	info, err := s.updaterClient.CheckNewVersion(ctx)
	if err != nil {
		return nil, err
	}
	return &VersionInfo{
		Name:             info.Data.Name,
		InstalledVersion: info.Data.InstalledVersion,
		AvailableVersion: info.Data.AvailableVersion,
		NeedForUpdate:    info.Data.NeedForUpdate,
	}, nil
}

func (s *Service) CheckNewVersionProcessing(ctx context.Context) (*VersionInfo, error) {
	info, err := s.processing.CheckNewVersionProcessing(ctx)
	if err != nil {
		return nil, err
	}
	return &VersionInfo{
		Name:             info.Name,
		InstalledVersion: info.InstalledVersion,
		AvailableVersion: info.AvailableVersion,
		NeedForUpdate:    info.NeedForUpdate,
	}, nil
}

func (s *Service) CheckApplicationVersions(ctx context.Context) (*ApplicationVersion, error) {
	var (
		backendVersion    *VersionInfo
		processingVersion *VersionInfo
	)

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		version, err := s.CheckNewVersionBackend(egCtx)
		if err != nil {
			return err
		}

		backendVersion = version
		return nil
	})

	eg.Go(func() error {
		version, err := s.CheckNewVersionProcessing(egCtx)
		if err != nil {
			return err
		}

		processingVersion = version
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return &ApplicationVersion{
		BackendVersion:    backendVersion,
		ProcessingVersion: processingVersion,
	}, nil
}

func (s *Service) UpdateBackend(ctx context.Context) error {
	err := s.updaterClient.Update(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateProcessing(ctx context.Context) error {
	err := s.processing.UpdateProcessing(ctx)
	if err != nil {
		return err
	}
	return nil
}
