package wallet

import (
	"context"
	"fmt"
	"time"

	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_update_balance_queue"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallet_addresses"

	"github.com/jackc/pgx/v5"
)

const defaultInterval = time.Second * 10

type BalanceUpdater interface {
	Run(ctx context.Context, tickerInterval time.Duration)
}

func (s *Service) Run(ctx context.Context, tickerInterval time.Duration) {
	ticker := time.NewTicker(min(tickerInterval, defaultInterval))
	for {
		select {
		case <-ticker.C:
			queue, err := s.storage.UpdateBalanceQueue().GetQueuedWithCurrency(ctx)
			if err != nil {
				s.logger.Errorw("failed to update balance queue", "error", err)
				continue
			}

			go s.processQueue(ctx, queue)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) processQueue(ctx context.Context, queue []*repo_update_balance_queue.GetQueuedWithCurrencyRow) {
	if !s.updateBalanceInProgress.CompareAndSwap(false, true) {
		return
	}
	defer s.updateBalanceInProgress.Store(false)

	for _, q := range queue {
		if err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
			if err := s.handleUpdateQueue(ctx, q, repos.WithTx(tx)); err != nil {
				return err
			}

			return s.storage.UpdateBalanceQueue(repos.WithTx(tx)).Delete(ctx, q.UpdateBalanceQueue.ID)
		}); err != nil {
			s.logger.Errorw("failed to handle update balance queue", "error", err)
			continue
		}
	}
}

func (s *Service) handleUpdateQueue(ctx context.Context, q *repo_update_balance_queue.GetQueuedWithCurrencyRow, opts repos.Option) error {
	nativeCurr, err := q.Currency.Blockchain.NativeCurrency()
	if err != nil {
		return fmt.Errorf("failed to get native currency: %w", err)
	}

	if q.UpdateBalanceQueue.NativeTokenBalanceUpdate {
		if err = s.storage.WalletAddresses(opts).UpdateWalletNativeTokenBalance(
			ctx,
			repo_wallet_addresses.UpdateWalletNativeTokenBalanceParams{
				Address:    q.UpdateBalanceQueue.Address,
				CurrencyID: nativeCurr,
				Blockchain: q.Currency.Blockchain.String(),
			},
		); err != nil {
			return fmt.Errorf("failed to update addresses for update balance queue: %w", err)
		}

		// Only native currency update is required
		if nativeCurr == q.Currency.ID {
			return nil
		}
	}

	err = s.storage.WalletAddresses(opts).UpdateWalletBalance(ctx, q.UpdateBalanceQueue.Address, q.UpdateBalanceQueue.CurrencyID)
	if err != nil {
		return fmt.Errorf("failed to update addresses for update balance queue: %w", err)
	}

	return nil
}
