package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_store_secrets"
	"github.com/dv-net/dv-merchant/internal/tools/str"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ISecret interface {
	GetSecretByStore(ctx context.Context, storeID uuid.UUID) (string, error)
	GenerateSecret(ctx context.Context, storeID uuid.UUID, opts ...repos.Option) (string, error)
	GenerateNewSecret(ctx context.Context, storeID uuid.UUID, opts ...repos.Option) (string, error)
	RemoveSecretByStore(ctx context.Context, storeID uuid.UUID) error
}

func (s *Service) GetSecretByStore(ctx context.Context, storeID uuid.UUID) (string, error) {
	secret, err := s.storage.StoreSecrets().GetSecretByStoreID(ctx, storeID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrStoreSecretNotFound
	}
	if err != nil {
		return "", fmt.Errorf("fetch secret by store: %w", err)
	}

	return secret, nil
}

func (s *Service) GenerateSecret(ctx context.Context, storeID uuid.UUID, opts ...repos.Option) (string, error) {
	secret, err := str.RandomString(40)
	if err != nil {
		return "", err
	}

	createdSecret, err := s.storage.StoreSecrets(opts...).Create(ctx, repo_store_secrets.CreateParams{
		StoreID: storeID,
		Secret:  secret,
	})
	if err != nil {
		return "", fmt.Errorf("create secret: %w", err)
	}

	return createdSecret.Secret, nil
}

func (s *Service) GenerateNewSecret(ctx context.Context, storeID uuid.UUID, opts ...repos.Option) (string, error) {
	secret, err := str.RandomString(40)
	if err != nil {
		return "", err
	}

	createdSecret, err := s.storage.StoreSecrets(opts...).UpdateSecret(ctx, repo_store_secrets.UpdateSecretParams{
		StoreID: storeID,
		Secret:  secret,
	})
	if err != nil {
		return "", fmt.Errorf("update secret: %w", err)
	}

	return createdSecret.Secret, nil
}

func (s *Service) RemoveSecretByStore(ctx context.Context, storeID uuid.UUID) error {
	return s.storage.StoreSecrets().DeleteBuStoreID(ctx, storeID)
}
