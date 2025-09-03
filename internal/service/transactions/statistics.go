package transactions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/transactions_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transactions"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) GetTransactionStats(ctx context.Context, user *models.User, params transactions_request.GetStatistics) ([]*repo_transactions.StatisticsRow, error) {
	var dateFrom *time.Time
	if params.DateFrom != nil {
		date, err := util.ParseDate(*params.DateFrom)
		if err != nil {
			return nil, err
		}

		dateFrom = date
	}

	var dateTo *time.Time
	if params.DateTo != nil {
		date, err := util.ParseDate(*params.DateTo)
		if err != nil {
			return nil, err
		}

		dateTo = date
	}

	userLocation, err := time.LoadLocation(user.Location)
	if err != nil {
		userLocation = time.UTC
	}

	res, err := s.storage.Transactions().GetStatistics(ctx, repo_transactions.GetStatisticsParams{
		UserID:         user.ID,
		TargetTimeZone: userLocation,
		Currencies:     params.Currencies,
		Resolution:     repo_transactions.Resolution(params.Resolution),
		Type:           params.Type,
		IsSystem:       params.IsSystem,
		MinAmountUSD:   params.MinAmountUSD,
		Blockchain:     params.Blockchain,
		DateFrom:       dateFrom,
		DateTo:         dateTo,
	})
	if err != nil {
		s.log.Error("fetch tx stats error", err)
		return nil, errors.New("prepare tx stats error")
	}

	return res, nil
}

func (s *Service) DepositStatistics(ctx context.Context, params StatisticsParams) ([]StatisticsDTO, error) {
	if params.Resolution == nil {
		return s.fetchDefaultStats(ctx, params.User.ID, params.StoreUUIDS, params.User.Location)
	}

	return s.fetchStatsByResolution(ctx, params)
}

func (s *Service) fetchStatsByResolution(
	ctx context.Context,
	params StatisticsParams,
) ([]StatisticsDTO, error) {
	if params.DateFrom == nil || params.DateTo == nil {
		return nil, fmt.Errorf("date_from and date_to must not be nil")
	}

	dateFrom, err := util.ParseDate(*params.DateFrom)
	if err != nil {
		return nil, fmt.Errorf("parse date_from: %w", err)
	}

	dateTo, err := util.ParseDate(*params.DateTo)
	if err != nil {
		return nil, fmt.Errorf("parse date_to: %w", err)
	}

	res, err := s.storage.Transactions().CalculateDepositStatistics(ctx, repo_transactions.CalculateDepositStatisticsParams{
		UserID:     params.User.ID,
		Resolution: string(*params.Resolution),
		DateFrom: pgtype.Timestamp{
			Time:  *dateFrom,
			Valid: true,
		},
		DateTo: pgtype.Timestamp{
			Time:  *dateTo,
			Valid: true,
		},
		StoreUuids:  params.StoreUUIDS,
		Timezone:    params.User.Location,
		CurrencyIds: params.CurrencyIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("calculate deposit statistics: %w", err)
	}

	preparedStats := make([]StatisticsDTO, 0, len(res))
	for _, stats := range res {
		var currencyTxCounts map[string]CurrencyDetailsDTO
		err := json.Unmarshal(stats.CurrencyStats, &currencyTxCounts)
		if err != nil {
			currencyTxCounts = make(map[string]CurrencyDetailsDTO)
		}
		preparedStats = append(preparedStats, StatisticsDTO{
			Date:              stats.Date.Time,
			Type:              StatisticsResolution(stats.Resolution),
			SumUsd:            stats.Sum.String(),
			TransactionsCount: stats.TxCount,
			DetailsByCurrency: currencyTxCounts,
		})
	}

	return preparedStats, nil
}

func (s *Service) fetchDefaultStats(
	ctx context.Context,
	userID uuid.UUID,
	storeUUIDs []uuid.UUID,
	userLocation string,
) ([]StatisticsDTO, error) {
	startTime := time.Now()
	currTime := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location())

	params := repo_transactions.CalculateDepositStatisticsParams{
		Resolution: string(StatisticsResolutionDay),
		UserID:     userID,
		DateFrom: pgtype.Timestamp{
			Time:             currTime.AddDate(0, 0, -5),
			InfinityModifier: 0,
			Valid:            true,
		},
		DateTo: pgtype.Timestamp{
			Time:             startTime,
			InfinityModifier: 0,
			Valid:            true,
		},
		StoreUuids: storeUUIDs,
		Timezone:   userLocation,
	}
	// last 5 days
	daysStats, err := s.storage.Transactions().CalculateDepositStatistics(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetch default stats: %w", err)
	}

	// last month stats
	startOfLastMonth := startTime.AddDate(0, -1, 0)

	totalParams := repo_transactions.CalculateDepositStatisticsTotalParams{
		UserID:     userID,
		StoreUuids: storeUUIDs,
		DateFrom: pgtype.Timestamp{
			Time:             startOfLastMonth,
			InfinityModifier: 0,
			Valid:            true,
		},
		DateTo: pgtype.Timestamp{
			Time:             startTime,
			InfinityModifier: 0,
			Valid:            true,
		},
		Timezone: userLocation,
	}

	summaryLength := len(daysStats)
	stats := make([]*repo_transactions.CalculateDepositStatisticsRow, 0, summaryLength)
	stats = append(stats, daysStats...)

	aggregatedResult := make([]StatisticsDTO, 0, len(stats))
	for _, val := range stats {
		var currencyTxCounts map[string]CurrencyDetailsDTO
		err := json.Unmarshal(val.CurrencyStats, &currencyTxCounts)
		if err != nil {
			currencyTxCounts = make(map[string]CurrencyDetailsDTO)
		}

		aggregatedResult = append(aggregatedResult, StatisticsDTO{
			Date:              val.Date.Time,
			Type:              StatisticsResolution(val.Resolution),
			SumUsd:            val.Sum.String(),
			TransactionsCount: val.TxCount,
			DetailsByCurrency: currencyTxCounts,
		})
	}

	lastMonthStats, err := s.storage.Transactions().CalculateDepositStatisticsTotal(ctx, totalParams)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("fetch default month stats: %w", err)
	}

	if lastMonthStats.TxCount > 0 && !errors.Is(err, pgx.ErrNoRows) {
		var currencyTxCounts map[string]CurrencyDetailsDTO
		err = json.Unmarshal(lastMonthStats.CurrencyStats, &currencyTxCounts)
		if err != nil {
			currencyTxCounts = make(map[string]CurrencyDetailsDTO)
		}
		aggregatedResult = append(aggregatedResult, StatisticsDTO{
			Date:              startOfLastMonth,
			Type:              StatisticsResolutionMonth,
			SumUsd:            lastMonthStats.Sum.String(),
			TransactionsCount: lastMonthStats.TxCount,
			DetailsByCurrency: currencyTxCounts,
		})
	}

	return aggregatedResult, nil
}
