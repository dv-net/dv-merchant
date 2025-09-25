package processing

import (
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/dv-net/dv-merchant/internal/dto"
	systemv1 "github.com/dv-net/dv-processing/api/processing/system/v1"
	"golang.org/x/net/context"
)

type IProcessingSystem interface {
	GetProcessingSystemInfo(ctx context.Context) (Info, error)
	UpdateProcessing(ctx context.Context) error
	CheckNewVersionProcessing(ctx context.Context) (*NewVersion, error)
	GetProcessingLogs(ctx context.Context) ([]dto.LogDTO, error)
}

var _ IProcessingSystem = (*Service)(nil)

func (s *Service) GetProcessingSystemInfo(ctx context.Context) (Info, error) {
	resp, err := s.processingService.System().Info(ctx, connect.NewRequest(&systemv1.InfoRequest{}))
	if err != nil {
		return Info{}, fmt.Errorf("fetch system info from processing: %w", err)
	}

	return Info{Version: resp.Msg.Version, Hash: resp.Msg.Commit}, nil
}

func (s *Service) UpdateProcessing(ctx context.Context) error {
	_, err := s.processingService.System().UpdateToNewVersion(ctx, connect.NewRequest(&systemv1.UpdateToNewVersionRequest{}))
	if err != nil {
		return fmt.Errorf("fetch system update to new version: %w", err)
	}
	return nil
}

func (s *Service) CheckNewVersionProcessing(ctx context.Context) (*NewVersion, error) {
	resp, err := s.processingService.System().CheckNewVersion(ctx, connect.NewRequest(&systemv1.CheckNewVersionRequest{}))
	if err != nil {
		return nil, fmt.Errorf("fetch system check to new version: %w", err)
	}
	return &NewVersion{
		Name:             resp.Msg.Name,
		InstalledVersion: resp.Msg.InstalledVersion,
		AvailableVersion: resp.Msg.AvailableVersion,
		NeedForUpdate:    resp.Msg.NeedForUpdate,
	}, nil
}

func (s *Service) GetProcessingLogs(ctx context.Context) ([]dto.LogDTO, error) {
	resp, err := s.processingService.System().GetLastLogs(ctx, connect.NewRequest(&systemv1.GetLastLogsRequest{}))
	if err != nil {
		return nil, fmt.Errorf("failed get processing logs: %w", err)
	}

	logs := make([]dto.LogDTO, 0, len(resp.Msg.Logs))
	for _, l := range resp.Msg.Logs {
		logs = append(logs, dto.LogDTO{
			Level:   l.Level,
			Message: l.Message,
			Time:    l.Time.AsTime().Format(time.RFC3339Nano),
		})
	}

	return logs, nil
}
