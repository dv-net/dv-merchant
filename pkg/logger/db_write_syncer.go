package logger

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_logs"

	"github.com/google/uuid"
)

type DBWriteSyncer struct {
	storageSvc storage.IStorage
}

type LogDTO struct {
	TypeSlug  string
	Level     string
	Message   string
	Status    string
	ProcessID uuid.UUID
}

func NewDBWriteSyncer(storageSvc storage.IStorage) *DBWriteSyncer {
	return &DBWriteSyncer{storageSvc: storageSvc}
}

func (w *DBWriteSyncer) WriteLog(ctx context.Context, logDTO LogDTO) (*models.Log, error) {
	return w.storageSvc.Logs().Create(ctx, repo_logs.CreateParams{
		LogTypeSlug: logDTO.TypeSlug,
		Level:       logDTO.Level,
		Message:     logDTO.Message,
		Status:      logDTO.Status,
		ProcessID:   logDTO.ProcessID,
	})
}
