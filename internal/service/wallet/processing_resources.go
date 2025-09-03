package wallet

import (
	"context"
	"fmt"

	"sync"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_tron_wallet_balance_statistics"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_users"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/dv-net/dv-merchant/pkg/retry"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

const MaxWorkers = 50
const MinUpdateInterval = time.Hour * 1

func (s *Service) ProcessingBalanceStatsInBackground(ctx context.Context, updateInterval time.Duration) {
	ticker := time.NewTicker(max(updateInterval, MinUpdateInterval))
	defer ticker.Stop()

	go func() {
		if err := s.processingBalanceStats(ctx); err != nil {
			s.logger.Error("update processing wallets stats", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go func() {
				if err := s.processingBalanceStats(ctx); err != nil {
					s.logger.Error("update processing wallets stats", err)
				}
			}()
		}
	}
}

func (s *Service) processingBalanceStats(ctx context.Context) error {
	if !s.updateProcessingStatsInProgress.CompareAndSwap(false, true) {
		return nil
	}
	defer s.updateProcessingStatsInProgress.Store(false)

	ownerIDs, err := s.storage.Users().GetActiveProcessingOwnersWithTronDelegate(ctx, repo_users.GetActiveProcessingOwnersWithTronDelegateParams{
		TronSettingName:  setting.TransferType,
		TronSettingValue: setting.TransferByResource.String(),
	})
	if err != nil {
		return err
	}

	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(MaxWorkers)

	for _, ownerID := range ownerIDs {
		eg.TryGo(func() error {
			return retry.New(
				retry.WithPolicy(retry.PolicyLinear),
				retry.WithMaxAttempts(3),
			).Do(func() error {
				wallets, err := s.processingService.GetOwnerProcessingWallets(egCtx, processing.GetOwnerProcessingWalletsParams{
					OwnerID:    ownerID.UUID,
					Blockchain: util.Pointer(models.BlockchainTron),
					Tiny:       util.Pointer(false),
				})
				if err != nil {
					return err
				}

				return s.CreateTronWalletBalances(egCtx, ownerID.UUID, wallets)
			})
		})
	}

	return eg.Wait()
}

func (s *Service) CreateTronWalletBalances(ctx context.Context, ownerID uuid.UUID, wallets []processing.WalletProcessing) error {
	params := make([]repo_tron_wallet_balance_statistics.InsertTronWalletBalanceStatisticsBatchParams, 0, len(wallets))
	for _, w := range wallets {
		if w.AdditionalData == nil || w.AdditionalData.TronData == nil {
			s.logger.Info("skip processing tron wallet resources - no data received from processing", "address", w.Address)
			continue
		}

		resources, err := s.extractTargetResources(w)
		if err != nil {
			s.logger.Error("extract processing resources", err)
			continue
		}

		params = append(params, repo_tron_wallet_balance_statistics.InsertTronWalletBalanceStatisticsBatchParams{
			ProcessingOwnerID:  ownerID,
			Address:            w.Address,
			StakedBandwidth:    resources.StakedBandwidth,
			StakedEnergy:       resources.StakedEnergy,
			DelegatedEnergy:    resources.DelegatedEnergy,
			DelegatedBandwidth: resources.DelegatedBandwidth,
			AvailableBandwidth: resources.AvailableBandwidth,
			AvailableEnergy:    resources.AvailableEnergy,
		})
	}

	if len(params) == 0 {
		return nil
	}

	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		errChan := make(chan error, len(params))
		batchRes := s.storage.TronWalletBalanceStatistics(repos.WithTx(tx)).InsertTronWalletBalanceStatisticsBatch(ctx, params)
		defer func() {
			if err := batchRes.Close(); err != nil {
				s.logger.Error("batch tron wallet balance stats close error", err)
			}
		}()

		wg := sync.WaitGroup{}
		wg.Add(len(params))

		batchRes.Exec(func(_ int, err error) {
			defer wg.Done()
			errChan <- err
		})

		go func() {
			wg.Wait()
			close(errChan)
		}()

		for err := range errChan {
			if err != nil {
				return fmt.Errorf("batch create transfer system transactions: %w", err)
			}
		}

		return nil
	})
}

type resourceInfo struct {
	// Staked owned TRX amount
	StakedBandwidth decimal.Decimal
	StakedEnergy    decimal.Decimal

	// Delegated from external accounts
	DelegatedBandwidth decimal.Decimal
	DelegatedEnergy    decimal.Decimal

	// Total  resources
	TotalEnergy    decimal.Decimal
	TotalBandwidth decimal.Decimal

	// Available  resources
	AvailableEnergy    decimal.Decimal
	AvailableBandwidth decimal.Decimal
}

func (s *Service) extractTargetResources(w processing.WalletProcessing) (*resourceInfo, error) {
	stakedBwth, err := decimal.NewFromString(w.AdditionalData.TronData.StackedBandwidth)
	if err != nil {
		return nil, fmt.Errorf("wallet: %s, parse processing wallet staked_bandwidth: %w", w.Address, err)
	}

	stakedEnergy, err := decimal.NewFromString(w.AdditionalData.TronData.StackedEnergy)
	if err != nil {
		return nil, fmt.Errorf("wallet: %s, parse processing wallet staked_energy: %w", w.Address, err)
	}

	totalBwth, err := decimal.NewFromString(w.AdditionalData.TronData.TotalBandwidth)
	if err != nil {
		return nil, fmt.Errorf("wallet: %s, parse processing wallet total_bandwidth: %w", w.Address, err)
	}

	totalEnergy, err := decimal.NewFromString(w.AdditionalData.TronData.TotalEnergy)
	if err != nil {
		return nil, fmt.Errorf("wallet: %s, parse processing wallet total_energy: %w", w.Address, err)
	}

	totalUsedBwth, err := decimal.NewFromString(w.AdditionalData.TronData.TotalUsedBandwidth)
	if err != nil {
		return nil, fmt.Errorf("wallet: %s, parse processing wallet total_bandwidth: %w", w.Address, err)
	}

	totalUsedEnergy, err := decimal.NewFromString(w.AdditionalData.TronData.TotalUsedEnergy)
	if err != nil {
		return nil, fmt.Errorf("wallet: %s, parse processing wallet total_energy: %w", w.Address, err)
	}

	// DelegatedBandwidth = TotalBandwidthUsed - StackedBandwidth
	delegatedBwth := totalBwth.Sub(stakedBwth)
	if delegatedBwth.LessThan(decimal.Zero) {
		s.logger.Warn("negative delegated bandwidth calculated", "address", w.Address, "total_bandwidth", totalBwth, "stacked_bandwidth", stakedBwth)
		delegatedBwth = decimal.Zero
	}

	// DelegatedEnergy = TotalEnergyUsed - StakedEnergy
	delegatedEnergy := totalEnergy.Sub(stakedEnergy)
	if delegatedEnergy.LessThan(decimal.Zero) {
		s.logger.Warn("negative delegated energy calculated", "address", w.Address, "total_energy", totalEnergy, "stacked_energy", stakedEnergy)
		delegatedEnergy = decimal.Zero
	}

	return &resourceInfo{
		StakedBandwidth:    stakedBwth,
		DelegatedBandwidth: delegatedBwth,
		StakedEnergy:       stakedEnergy,
		DelegatedEnergy:    delegatedEnergy,
		AvailableEnergy:    totalEnergy.Sub(totalUsedEnergy),
		AvailableBandwidth: totalBwth.Sub(totalUsedBwth),
	}, nil
}
