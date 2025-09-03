package transactions

import (
	"context"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_unconfirmed_transactions"
)

type IUnconfirmedTransaction interface {
	CreateUnconfirmedTransaction(ctx context.Context, params repo_unconfirmed_transactions.CreateParams, opts ...repos.Option) (*models.UnconfirmedTransaction, error)
	GetUnconfirmedByHash(ctx context.Context, hash, blockchain string, opts ...repos.Option) (*models.UnconfirmedTransaction, error)
}

type IUnconfirmedTransactionCollapser interface {
	Run(ctx context.Context, collapseInterval time.Duration)
}

func (s *Service) Run(ctx context.Context, collapseInterval time.Duration) {
	ticker := time.NewTicker(collapseInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := s.storage.UnconfirmedTransactions().CollapseAllByConfirmed(ctx)
			if err != nil {
				s.log.Error("collapse unconfirmed transaction failed", err)
			}
		case <-ctx.Done():
			s.log.Info("unconfirmed collapser finished by ctx")
			return
		}
	}
}

func (s *Service) GetUnconfirmedByHash(ctx context.Context, hash, blockchain string, opts ...repos.Option) (*models.UnconfirmedTransaction, error) {
	unconfirmedTransaction, err := s.storage.UnconfirmedTransactions(opts...).GetOneByHashAndBlockchain(
		ctx,
		repo_unconfirmed_transactions.GetOneByHashAndBlockchainParams{
			TxHash:     hash,
			Blockchain: blockchain,
		},
	)
	if err != nil {
		return nil, err
	}
	return unconfirmedTransaction, nil
}

func (s *Service) CreateUnconfirmedTransaction(ctx context.Context, params repo_unconfirmed_transactions.CreateParams, opts ...repos.Option) (*models.UnconfirmedTransaction, error) {
	unconfirmedTransaction, err := s.storage.UnconfirmedTransactions(opts...).Create(ctx, params)
	if err != nil {
		return nil, err
	}
	return unconfirmedTransaction, nil
}
