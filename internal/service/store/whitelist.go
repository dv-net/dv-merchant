package store

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_store_whitelist"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type IStoreWhitelist interface {
	GetStoreWhitelist(ctx context.Context, storeID uuid.UUID) ([]*models.StoreWhitelist, error)
	PatchStoreWhitelist(ctx context.Context, storeID uuid.UUID, ip string) ([]*models.StoreWhitelist, error)
	CreateStoreWhitelist(ctx context.Context, storeID uuid.UUID, ip []string) ([]*models.StoreWhitelist, error)
	DeleteStoreWhitelist(ctx context.Context, storeID uuid.UUID, opts ...repos.Option) error
	DeleteStoreWhitelistByIP(ctx context.Context, storeID uuid.UUID, ip string, opts ...repos.Option) error
}

func (s *Service) GetStoreWhitelist(ctx context.Context, storeID uuid.UUID) ([]*models.StoreWhitelist, error) {
	storeWhitelist, err := s.storage.StoreWhitelist().Find(ctx, storeID)
	if err != nil {
		return nil, err
	}
	return storeWhitelist, nil
}

func (s *Service) PatchStoreWhitelist(ctx context.Context, storeID uuid.UUID, ip string) ([]*models.StoreWhitelist, error) {
	var storeWhitelists []*models.StoreWhitelist
	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		arg := repo_store_whitelist.CheckExistsByIPParams{
			StoreID: storeID,
			Ip:      ip,
		}
		exists, err := s.storage.StoreWhitelist().CheckExistsByIP(ctx, arg)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("store whitelist ip already exists")
		}

		params := repo_store_whitelist.CreateParams{
			StoreID: storeID,
			Ip:      ip,
		}
		storeWhitelist, err := s.storage.StoreWhitelist(repos.WithTx(tx)).Create(ctx, params)
		if err != nil {
			return err
		}
		storeWhitelists = append(storeWhitelists, storeWhitelist)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("transaction error: %w", err)
	}
	return storeWhitelists, nil
}

func (s *Service) CreateStoreWhitelist(ctx context.Context, storeID uuid.UUID, ip []string) ([]*models.StoreWhitelist, error) {
	var storeWhitelists []*models.StoreWhitelist
	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		err := s.DeleteStoreWhitelist(ctx, storeID, repos.WithTx(tx))
		if err != nil {
			return err
		}

		for _, ip := range ip {
			params := repo_store_whitelist.CreateParams{
				StoreID: storeID,
				Ip:      ip,
			}
			storeWhitelist, err := s.storage.StoreWhitelist(repos.WithTx(tx)).Create(ctx, params)
			if err != nil {
				return err
			}
			storeWhitelists = append(storeWhitelists, storeWhitelist)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("transaction error: %w", err)
	}
	return storeWhitelists, nil
}

func (s *Service) DeleteStoreWhitelist(ctx context.Context, storeID uuid.UUID, opts ...repos.Option) error {
	err := s.storage.StoreWhitelist(opts...).Delete(ctx, storeID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) DeleteStoreWhitelistByIP(ctx context.Context, storeID uuid.UUID, ip string, opts ...repos.Option) error {
	arg := repo_store_whitelist.DeleteByIPParams{
		StoreID: storeID,
		Ip:      ip,
	}

	err := s.storage.StoreWhitelist(opts...).DeleteByIP(ctx, arg)
	if err != nil {
		return err
	}
	return nil
}
