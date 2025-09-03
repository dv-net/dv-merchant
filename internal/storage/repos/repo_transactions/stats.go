package repo_transactions

import (
	"context"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/shopspring/decimal"

	"github.com/dv-net/dv-merchant/internal/models"
)

const defaultTxStatsTimeZone = "UTC"

type StatisticsRow struct {
	Date        string          `db:"date" json:"date"`
	AmountUSD   decimal.Decimal `db:"amount_usd" json:"amount_usd"`
	CurrencyIDs []string        `db:"currency_ids" json:"currency_ids"` //nolint:tagliatelle
	Resolution  Resolution      `db:"resolution" json:"resolution"`
}

type Resolution string

const (
	ResolutionHour Resolution = "hour"
	ResolutionDay  Resolution = "day"
)

type GetStatisticsParams struct {
	UserID         uuid.UUID
	TargetTimeZone *time.Location
	Currencies     []string                 `json:"currencies" query:"currencies"`
	Resolution     Resolution               `json:"resolution" query:"resolution" validate:"required,oneof=hour day"`
	Type           *models.TransactionsType `json:"type" query:"type" validate:"omitempty,oneof=deposit transfer" enums:"deposit,transfer"`
	IsSystem       bool                     `json:"is_system" query:"is_system"`
	MinAmountUSD   decimal.Decimal          `json:"min_amount_usd" query:"min_amount_usd"`
	Blockchain     *models.Blockchain       `json:"blockchain" query:"blockchain"`
	DateFrom       *time.Time               `json:"date_from" query:"date_from"`
	DateTo         *time.Time               `json:"date_to" query:"date_to"`
}

func (s *CustomQuerier) GetStatistics(
	ctx context.Context,
	params GetStatisticsParams,
) ([]*StatisticsRow, error) {
	timeZoneArg := params.TargetTimeZone.String()
	if timeZoneArg == "" {
		timeZoneArg = defaultTxStatsTimeZone
	}

	dateCond := fmt.Sprintf(
		"TO_CHAR(date_trunc('%s', TO_TIMESTAMP(transactions.created_at_index / 1000) AT TIME ZONE '%s'), 'YYYY-MM-DD HH24:MI')",
		params.Resolution,
		timeZoneArg,
	)

	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select(
		fmt.Sprintf("%s AS \"date\"", dateCond),
		"SUM(COALESCE(transactions.amount_usd, 0)) AS \"amount_usd\"",
		fmt.Sprintf("'%s' AS \"resolution\"", params.Resolution),
		"array_agg(distinct transactions.currency_id) AS \"currency_ids\"",
	).
		From("transactions").
		Where(sb.Equal("transactions.user_id", params.UserID.String())).
		GroupBy(dateCond).
		Having("SUM(COALESCE(transactions.amount_usd, 0)) > 0").
		OrderBy(dateCond + " ASC")

	// Add JOIN with currencies table only if Currencies or Blockchain filters are specified
	if len(params.Currencies) > 0 || params.Blockchain != nil {
		sb.JoinWithOption("INNER", "currencies", "currencies.id = transactions.currency_id")
	}

	if len(params.Currencies) > 0 {
		currencies := make([]any, len(params.Currencies))
		for i, id := range params.Currencies {
			currencies[i] = id
		}
		sb.Where(sb.In("transactions.currency_id", currencies...))
	}

	if params.Type != nil {
		sb.Where(sb.Equal("transactions.type", string(*params.Type)))
	}

	if !params.MinAmountUSD.IsZero() {
		sb.Where(sb.GreaterEqualThan("transactions.amount_usd", params.MinAmountUSD))
	}

	if params.DateFrom != nil {
		from := params.DateFrom.In(params.TargetTimeZone).UTC().Unix() * 1000
		sb.Where(sb.GreaterEqualThan("transactions.created_at_index", from))
	}

	if params.DateTo != nil {
		to := params.DateTo.In(params.TargetTimeZone).UTC().Unix() * 1000
		sb.Where(sb.LessThan("transactions.created_at_index", to))
	}

	if params.Blockchain != nil {
		sb.Where(sb.Equal("currencies.blockchain", string(*params.Blockchain)))
	}

	sb.Where(sb.Equal("transactions.is_system", params.IsSystem))

	sql, sqlArgs := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	items := make([]*StatisticsRow, 0)
	if err := pgxscan.Select(ctx, s.psql, &items, sql, sqlArgs...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	return items, nil
}
