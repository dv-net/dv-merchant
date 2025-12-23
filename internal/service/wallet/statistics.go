package wallet

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/exrate"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transfer_transactions"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_tron_wallet_balance_statistics"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallets"
	"github.com/dv-net/dv-merchant/internal/util"

	"github.com/dv-net/dv-processing/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

func (s *Service) SummarizeUserWalletsByCurrency(ctx context.Context, userID uuid.UUID, rates *exrate.Rates, minBalance decimal.Decimal) ([]SummaryDTO, error) {
	params := repo_wallets.SummarizeByCurrencyParams{
		UserID:     userID,
		Rates:      rates.Rate,
		IDs:        rates.CurrencyIDs,
		MinBalance: minBalance,
	}

	res, err := s.storage.Wallets().SummarizeByCurrency(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetch wallet summary: %w", err)
	}

	existingCurrencies := make(map[string]SummaryDTO, len(res))

	// Replace zeroed items with actual data
	for _, val := range res {
		existingCurrencies[val.ID] = SummaryDTO{
			Currency: CurrencyDTO{
				ID:              val.ID,
				Code:            val.Code,
				Name:            val.Name,
				Blockchain:      *val.Blockchain,
				SortOrder:       val.SortOrder,
				IsNative:        val.IsNative,
				ContractAddress: val.ContractAddress,
			},
			Balance:          val.Balance.String(),
			BalanceUSD:       val.AmountUsd.String(),
			Count:            val.Count,
			CountWithBalance: val.CountWithBalance,
		}
	}

	// Convert map to slice
	preparedRes := lo.MapToSlice[string, SummaryDTO](existingCurrencies, func(_ string, value SummaryDTO) SummaryDTO {
		return value
	})
	// sort first -> amountUSD > currencyID
	sort.Slice(preparedRes, func(i, j int) bool {
		amountI, _ := decimal.NewFromString(preparedRes[i].BalanceUSD)
		amountJ, _ := decimal.NewFromString(preparedRes[j].BalanceUSD)

		if !amountI.Equal(amountJ) {
			return amountI.GreaterThan(amountJ)
		}

		return preparedRes[i].Currency.ID < preparedRes[j].Currency.ID
	})

	return preparedRes, nil
}

// FetchTronResourceStatistics retrieves aggregated Tron wallet and transfer statistics for a user.
func (s *Service) FetchTronResourceStatistics(ctx context.Context, user *models.User, params FetchTronStatisticsParams) (map[string]CombinedStats, error) {
	if user == nil || !user.ProcessingOwnerID.Valid {
		return nil, fmt.Errorf("processing owner is undefined")
	}

	// Validate resolution
	resolution := params.Resolution
	if resolution == "" {
		resolution = FetchTronStatsResolutionDay // Default to day if not specified
	}

	dateFrom, dateTo, err := prepareDateRange(params)
	if err != nil {
		return nil, err
	}

	userLocation, err := time.LoadLocation(user.Location)
	if err != nil {
		s.logger.Warnw("invalid user location, fallback to UTC", "error", err)
		userLocation = time.UTC
	}

	balanceStats, err := s.storage.TronWalletBalanceStatistics().ApproximateByResolution(ctx, repo_tron_wallet_balance_statistics.ApproximateByResolutionParams{
		Timezone:        userLocation.String(),
		DateFrom:        dateFrom,
		DateTo:          dateTo,
		ProcessingOwner: user.ProcessingOwnerID.UUID,
		Resolution:      resolution,
	})
	if err != nil {
		s.logger.Errorw("failed to fetch tron wallet balance statistics", "error", err)
		return nil, fmt.Errorf("failed to fetch tron wallet balance statistics: %w", err)
	}

	transferStats, err := s.storage.TransferTransactions().CalculateTransfersExpense(ctx, repo_transfer_transactions.CalculateTransfersExpenseParams{
		UserID:     user.ID,
		Timezone:   userLocation.String(),
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		Blockchain: utils.Pointer(models.BlockchainTron),
		Resolution: resolution,
		TxTypes: []string{
			models.TransferTransactionTypeTransfer.String(),
			models.TransferTransactionTypeReclaimResources.String(),
			models.TransferTransactionTypeDelegateResources.String(),
		},
	})
	if err != nil {
		s.logger.Errorw("failed to fetch transfer expenses", "error", err)
		return nil, fmt.Errorf("failed to fetch transfer expenses: %w", err)
	}

	// Combine statistics by date
	result := make(map[string]CombinedStats, len(balanceStats)+len(transferStats))
	for _, stat := range balanceStats {
		dateKey := formatDateKey(stat.Day.Time, resolution)
		combined := result[dateKey]
		combined.StakedBandwidth = combined.StakedBandwidth.Add(stat.StakedBandwidth)
		combined.StakedEnergy = combined.StakedEnergy.Add(stat.StakedEnergy)
		combined.DelegatedEnergy = combined.DelegatedEnergy.Add(stat.DelegatedEnergy)
		combined.DelegatedBandwidth = combined.DelegatedBandwidth.Add(stat.DelegatedBandwidth)
		combined.AvailableBandwidth = combined.AvailableBandwidth.Add(stat.AvailableBandwidth)
		combined.AvailableEnergy = combined.AvailableEnergy.Add(stat.AvailableEnergy)
		result[dateKey] = combined
	}

	for _, expense := range transferStats {
		dateKey := formatDateKey(expense.Day.Time, resolution)
		combined := result[dateKey]
		combined.TransferCount += expense.TransfersCount
		combined.TotalTrxFee = combined.TotalTrxFee.Add(expense.TotalTrxFee)
		combined.TotalBandwidthUsed = combined.TotalBandwidthUsed.Add(expense.TotalBandwidth)
		combined.TotalEnergyUsed = combined.TotalEnergyUsed.Add(expense.TotalEnergy)
		result[dateKey] = combined
	}

	return result, nil
}

// formatDateKey formats the date key based on the resolution
func formatDateKey(t time.Time, resolution string) string {
	switch resolution {
	case FetchTronStatsResolutionHour:
		return t.Format("2006-01-02 15:00")
	default:
		return t.Format("2006-01-02")
	}
}

func prepareDateRange(params FetchTronStatisticsParams) (pgtype.Timestamp, pgtype.Timestamp, error) {
	now := utils.Pointer(time.Now().UTC())
	defaultTo := now
	defaultFrom := utils.Pointer(now.AddDate(0, 0, -7))

	from := defaultFrom
	if params.DateFrom != nil {
		var err error
		from, err = util.ParseDate(*params.DateFrom)
		if err != nil {
			return pgtype.Timestamp{}, pgtype.Timestamp{}, fmt.Errorf("invalid date_from: %w", err)
		}
	}

	to := defaultTo
	if params.DateTo != nil {
		var err error
		to, err = util.ParseDate(*params.DateTo)
		if err != nil {
			return pgtype.Timestamp{}, pgtype.Timestamp{}, fmt.Errorf("invalid date_to: %w", err)
		}
	}

	if from.After(*to) {
		return pgtype.Timestamp{}, pgtype.Timestamp{}, fmt.Errorf("date_from must be before date_to")
	}

	return pgtype.Timestamp{Time: from.UTC(), Valid: true},
		pgtype.Timestamp{Time: to.UTC(), Valid: true},
		nil
}
