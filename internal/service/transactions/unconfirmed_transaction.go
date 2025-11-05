package transactions

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_unconfirmed_transactions"
	transactionsv2 "github.com/dv-net/dv-proto/gen/go/eproxy/transactions/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type IUnconfirmedTransaction interface {
	CreateUnconfirmedTransaction(ctx context.Context, params repo_unconfirmed_transactions.CreateParams, opts ...repos.Option) (*models.UnconfirmedTransaction, error)
	GetUnconfirmedByHash(ctx context.Context, hash string, blockchain models.Blockchain, opts ...repos.Option) (*models.UnconfirmedTransaction, error)
	GetUnconfirmedTransactions(ctx context.Context, userID uuid.UUID, txType models.TransactionsType) ([]*models.UnconfirmedTransaction, error)
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
			err := s.storage.UnconfirmedTransactions().CollapseAllByConfirmedDeposit(ctx)
			if err != nil {
				s.log.Errorw("collapse unconfirmed transaction failed", "error", err)
			}

			err = s.CollapseTransfers(ctx)
			if err != nil {
				s.log.Errorw("collapse unconfirmed transfers failed", "error", err)
			}
		case <-ctx.Done():
			s.log.Info("unconfirmed collapser finished by ctx")
			return
		}
	}
}

func (s *Service) GetUnconfirmedByHash(
	ctx context.Context,
	hash string,
	blockchain models.Blockchain,
	opts ...repos.Option,
) (*models.UnconfirmedTransaction, error) {
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

func (s *Service) CollapseTransfers(ctx context.Context) error {
	uTransactions, err := s.storage.UnconfirmedTransactions().GetAllTransfer(ctx)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	for _, uTx := range uTransactions {
		blockchain, err := uTx.Blockchain.ToEPb()
		if err != nil {
			s.log.Errorw("failed to convert blockchain to eproxy pb", "blockchain", uTx.Blockchain, "error", err)
			continue
		}
		txInfo, err := s.epr.Transactions().GetInfo(ctx, connect.NewRequest(&transactionsv2.GetInfoRequest{
			Hash:       uTx.TxHash,
			Blockchain: blockchain,
		}))

		if err != nil {
			s.log.Errorw("failed to get transaction info from eproxy", "hash", uTx.TxHash, "blockchain", uTx.Blockchain, "error", err)
			continue
		}

		if txInfo.Msg.Transaction.Confirmations > uTx.Blockchain.ConfirmationBlockCount() {
			err = s.storage.UnconfirmedTransactions().DeleteByTxHash(ctx, uTx.ID)
			if err != nil {
				s.log.Errorw("failed to delete unconfirmed transaction", "hash", uTx.TxHash, "error", err)
				continue
			}
			s.log.Infow("unconfirmed transaction collapsed", "hash", uTx.TxHash)
		}
	}
	return nil
}

func (s *Service) GetUnconfirmedTransactions(ctx context.Context, userID uuid.UUID, txType models.TransactionsType) ([]*models.UnconfirmedTransaction, error) {
	uTxs, err := s.storage.UnconfirmedTransactions().GetByType(ctx, repo_unconfirmed_transactions.GetByTypeParams{
		Type:   txType,
		UserID: userID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	return uTxs, nil
}
