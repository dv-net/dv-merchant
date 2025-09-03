package store

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_store_api_keys"
	"github.com/dv-net/dv-merchant/internal/tools/str"

	"github.com/google/uuid"
)

type IStoreAPIKey interface {
	GetStoreAPIKeyByID(ctx context.Context, ID uuid.UUID) (*models.StoreApiKey, error)
	GetAPIKeyByStoreID(ctx context.Context, storeID uuid.UUID) ([]*models.StoreApiKey, error)
	CreateAPIKey(ctx context.Context, store *models.Store, opts ...repos.Option) (*models.StoreApiKey, error)
	GenerateAPIKey(ctx context.Context, storeID uuid.UUID, opts ...repos.Option) (*models.StoreApiKey, error)
	UpdateStatusStoreAPIKey(ctx context.Context, ID uuid.UUID, status bool, opts ...repos.Option) (*models.StoreApiKey, error)
	DeleteAPIKey(ctx context.Context, ID uuid.UUID) error
}

func (s *Service) GetStoreAPIKeyByID(ctx context.Context, id uuid.UUID) (*models.StoreApiKey, error) {
	storeAPIKey, err := s.storage.StoreAPIKeys().GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	return storeAPIKey, nil
}

func (s *Service) GetAPIKeyByStoreID(ctx context.Context, storeID uuid.UUID) ([]*models.StoreApiKey, error) {
	storeAPIKeys, err := s.storage.StoreAPIKeys().GetByStoreId(ctx, storeID)
	if err != nil {
		return nil, err
	}
	return storeAPIKeys, nil
}

func (s *Service) CreateAPIKey(ctx context.Context, store *models.Store, opts ...repos.Option) (*models.StoreApiKey, error) {
	key, err := str.RandomString(64)
	if err != nil {
		return nil, err
	}
	params := repo_store_api_keys.CreateParams{
		Key:     key,
		StoreID: store.ID,
		Enabled: true,
	}

	storeAPIKey, err := s.storage.StoreAPIKeys(opts...).Create(ctx, params)
	if err != nil {
		return nil, err
	}
	return storeAPIKey, nil
}

func (s *Service) GenerateAPIKey(ctx context.Context, storeID uuid.UUID, opts ...repos.Option) (*models.StoreApiKey, error) {
	key, err := str.RandomString(64)
	if err != nil {
		return nil, err
	}
	params := repo_store_api_keys.UpdateKeyParams{
		Key:     key,
		StoreID: storeID,
	}

	storeAPIKey, err := s.storage.StoreAPIKeys(opts...).UpdateKey(ctx, params)
	if err != nil {
		return nil, err
	}
	return storeAPIKey, nil
}

func (s *Service) UpdateStatusStoreAPIKey(ctx context.Context, id uuid.UUID, status bool, opts ...repos.Option) (*models.StoreApiKey, error) {
	params := repo_store_api_keys.UpdateStatusParams{
		Enabled: status,
		ID:      id,
	}

	storeAPIKey, err := s.storage.StoreAPIKeys(opts...).UpdateStatus(ctx, params)
	if err != nil {
		return nil, err
	}
	return storeAPIKey, nil
}

func (s *Service) DeleteAPIKey(ctx context.Context, id uuid.UUID) error {
	err := s.storage.StoreAPIKeys().Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
