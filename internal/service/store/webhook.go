package store

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/store_webhook_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_store_webhooks"

	"github.com/google/uuid"
)

type IStoreWebhooks interface {
	GetStoreWebhookByID(ctx context.Context, ID uuid.UUID) (*models.StoreWebhook, error)
	GetStoreWebhookByStoreID(ctx context.Context, storeID uuid.UUID) ([]*models.StoreWebhook, error)
	CreateStoreWebhooks(ctx context.Context, store *models.Store, dto *store_webhook_request.CreateRequest, opts ...repos.Option) (*models.StoreWebhook, error)
	UpdateStoreWebhooks(ctx context.Context, ID uuid.UUID, dto *store_webhook_request.UpdateRequest, opts ...repos.Option) (*models.StoreWebhook, error)
	DeleteStoreWebhooks(ctx context.Context, ID uuid.UUID, opts ...repos.Option) error
}

func (s *Service) GetStoreWebhookByID(ctx context.Context, id uuid.UUID) (*models.StoreWebhook, error) {
	storeAPIKey, err := s.storage.StoreWebhooks().GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	return storeAPIKey, nil
}

func (s *Service) CreateStoreWebhooks(ctx context.Context, store *models.Store, dto *store_webhook_request.CreateRequest, opts ...repos.Option) (*models.StoreWebhook, error) {
	params := repo_store_webhooks.CreateParams{
		Url:     dto.URL,
		StoreID: store.ID,
		Enabled: dto.Enabled,
		Events:  dto.Events,
	}
	storeWebhook, err := s.storage.StoreWebhooks(opts...).Create(ctx, params)
	if err != nil {
		return nil, err
	}
	return storeWebhook, nil
}

func (s *Service) UpdateStoreWebhooks(ctx context.Context, id uuid.UUID, dto *store_webhook_request.UpdateRequest, opts ...repos.Option) (*models.StoreWebhook, error) {
	params := repo_store_webhooks.UpdateParams{
		Url:     dto.URL,
		Enabled: dto.Enabled,
		Events:  dto.Events,
		ID:      id,
	}
	storeWebhook, err := s.storage.StoreWebhooks(opts...).Update(ctx, params)
	if err != nil {
		return nil, err
	}
	return storeWebhook, nil
}

func (s *Service) GetStoreWebhookByStoreID(ctx context.Context, storeID uuid.UUID) ([]*models.StoreWebhook, error) {
	storeWebhooks, err := s.storage.StoreWebhooks().GetByStoreId(ctx, storeID)
	if err != nil {
		return nil, err
	}
	return storeWebhooks, nil
}

func (s *Service) DeleteStoreWebhooks(ctx context.Context, id uuid.UUID, opts ...repos.Option) error {
	err := s.storage.StoreWebhooks(opts...).Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
