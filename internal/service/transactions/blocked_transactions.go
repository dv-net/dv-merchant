package transactions

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_blocked_transactions"
	"github.com/dv-net/dv-merchant/pkg/dbutils/pgerror"
)

type IBlockedTransaction interface {
	CreateBlockedTransaction(ctx context.Context, dto CreateBlockedTransactionDTO, opts ...repos.Option) (*models.BlockedTransaction, error)
}

func (s *Service) CreateBlockedTransaction(ctx context.Context, dto CreateBlockedTransactionDTO, opts ...repos.Option) (*models.BlockedTransaction, error) {
	bTransaction, err := s.storage.BlockedTransactions(opts...).Create(ctx, repo_blocked_transactions.CreateParams{
		UserID:        dto.UserID,
		StoreID:       dto.StoreID,
		TransactionID: dto.TransactionID,
		AmlCheckID:    dto.AmlCheckID,
		WalletID:      dto.WalletID,
		RiskLevel:     dto.RiskLevel,
		Score:         dto.Score,
	})
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.log.Debug("error create blocked transactions", parsedErr)
		return nil, parsedErr
	}
	return bTransaction, nil
}
