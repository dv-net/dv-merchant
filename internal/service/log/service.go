package log

import (
	"context"
	"fmt"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_logs"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
)

type ILogService interface {
	StartProcess(ctx context.Context, slug string) (*Process, error)
	StopProcess(ctx context.Context, processID uuid.UUID) error
	GetProcess(ctx context.Context, uuid uuid.UUID) (*models.Log, error)
	GetLogsBySlug(ctx context.Context, slug string) ([]*InfoDTO, error)
	GetAllTypes(ctx context.Context) ([]*models.LogType, error)
}

type InfoDTO struct {
	ProcessID uuid.UUID                `json:"process_id"`
	Failure   bool                     `json:"failure"`
	CreatedAt time.Time                `json:"created_at"`
	Messages  []map[string]interface{} `json:"messages"`
}
type Service struct {
	storageSvc storage.IStorage
}

func NewService(storage storage.IStorage) *Service {
	return &Service{
		storageSvc: storage,
	}
}

type Process struct {
	ID       uuid.UUID
	TypeSlug string
}

func (s *Service) StartProcess(ctx context.Context, slug string) (*Process, error) {
	typeData, err := s.storageSvc.LogTypes().GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	newProcess := &Process{
		ID:       uuid.New(),
		TypeSlug: typeData.Slug,
	}
	return newProcess, nil
}

func (s Service) GetProcess(ctx context.Context, uuid uuid.UUID) (*models.Log, error) {
	process, err := s.storageSvc.Logs().GetByProcessID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return process, nil
}

func (s Service) GetLogsBySlug(ctx context.Context, slug string) ([]*InfoDTO, error) {
	logs, err := s.storageSvc.Logs().GetLogsBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	result := make([]*InfoDTO, 0, len(logs))
	for _, log := range logs {
		var messages []map[string]interface{}
		if err := json.Unmarshal(log.Messages, &messages); err != nil {
			return nil, fmt.Errorf("failed to decode messages for process ID %s: %w", log.ProcessID, err)
		}
		createdAt, ok := log.CreatedAt.(time.Time)
		if !ok {
			return nil, fmt.Errorf("unexpected type for CreatedAt in process ID %s: %T", log.ProcessID, log.CreatedAt)
		}
		dto := &InfoDTO{
			ProcessID: log.ProcessID,
			Failure:   log.Failure,
			CreatedAt: createdAt,
			Messages:  messages,
		}

		result = append(result, dto)
	}
	return result, nil
}

func (s Service) GetAllTypes(ctx context.Context) ([]*models.LogType, error) {
	logsTypes, err := s.storageSvc.LogTypes().GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return logsTypes, nil
}

func (s *Service) StopProcess(ctx context.Context, processID uuid.UUID) error {
	process, err := s.GetProcess(ctx, processID)
	if err != nil {
		return err
	}
	if process.Status == logger.Completed.String() {
		err := s.storageSvc.Logs().DeleteLogBySlug(ctx, repo_logs.DeleteLogBySlugParams{
			LogTypeSlug: process.LogTypeSlug,
			ProcessID:   processID,
		})
		if err != nil {
			return err
		}
		_, err = s.storageSvc.Logs().Create(ctx, repo_logs.CreateParams{
			Level:       "INFO",
			LogTypeSlug: process.LogTypeSlug,
			Message:     "Process completed successfully",
			Status:      logger.Completed.String(),
			ProcessID:   process.ID,
		})
		if err != nil {
			return err
		}
	} else {
		_, err := s.storageSvc.Logs().Create(ctx, repo_logs.CreateParams{
			Level:       "ERROR",
			LogTypeSlug: process.LogTypeSlug,
			Message:     "Process failed",
			Status:      logger.Failed.String(),
			ProcessID:   process.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
