package repo_aml_checks

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/pkg/dbutils"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/shopspring/decimal"
)

type GetByUserParams struct {
	storecmn.CommonFindParams
	ServiceSlug *models.AMLSlug
	Status      *models.AMLCheckStatus
	DateFrom    *time.Time
	DateTo      *time.Time
}

type AmlCheckDTO struct {
	models.AmlCheck
	Slug          models.AMLSlug      `json:"slug" db:"slug"`
	TxHash        *string             `json:"tx_hash" db:"tx_hash"`
	TxAmount      decimal.NullDecimal `json:"tx_amount" db:"tx_amount"`
	TxAmountUsd   decimal.NullDecimal `json:"tx_amount_usd" db:"tx_amount_usd"`
	TxCurrencyID  *string             `json:"tx_currency_id" db:"tx_currency_id"`
	CustomerEmail *string             `json:"customer_email" db:"customer_email"`
}

type FindRow struct {
	AmlCheckDTO
	History []*models.AmlCheckHistory
}

const maxLimit = 1000

func (s *CustomQuerier) GetByUser(ctx context.Context, usr *models.User, params GetByUserParams) (*storecmn.FindResponseWithFullPagination[*FindRow], error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select(
		`aml_checks.*`,
		`aml_services.slug`,
		`transactions.tx_hash AS tx_hash`,
		`transactions.amount AS tx_amount`,
		`transactions.amount_usd AS tx_amount_usd`,
		`transactions.currency_id AS tx_currency_id`,
		`wallets.email AS customer_email`,
	).
		From("aml_checks").
		JoinWithOption("INNER", "aml_services", "aml_services.id = aml_checks.service_id").
		JoinWithOption("LEFT", "transactions", "transactions.id = aml_checks.transaction_id").
		JoinWithOption("LEFT", "wallets", "wallets.id = transactions.wallet_id")

	countSb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	countSb.Select("COUNT(aml_checks.id)").
		From("aml_checks").
		JoinWithOption("INNER", "aml_services", "aml_services.id = aml_checks.service_id")

	timezone := time.UTC.String()
	if usr.Location != "" {
		if loc, err := time.LoadLocation(usr.Location); err == nil {
			timezone = loc.String()
		}
	}

	// the same filters for the select and count queries, so the pagination matches the items
	for _, b := range []*sqlbuilder.SelectBuilder{sb, countSb} {
		b.Where(b.Equal("aml_checks.user_id", usr.ID.String()))

		if params.ServiceSlug != nil {
			b.Where(b.Equal("aml_services.slug", params.ServiceSlug.String()))
		}

		if params.Status != nil {
			b.Where(b.Equal("aml_checks.status", params.Status.String()))
		}

		if params.DateFrom != nil {
			from := params.DateFrom.UTC().Format(time.RFC3339Nano)
			b.Where(b.GreaterEqualThan(fmt.Sprintf("aml_checks.created_at AT TIME ZONE '%s'", timezone), from))
		}

		if params.DateTo != nil {
			to := params.DateTo.UTC().Format(time.RFC3339Nano)
			b.Where(b.LessEqualThan(fmt.Sprintf("aml_checks.created_at AT TIME ZONE '%s'", timezone), to))
		}
	}

	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize, dbutils.WithMaxLimit(maxLimit))
	if err != nil {
		return nil, err
	}

	if params.OrderBy == "" {
		params.OrderBy = "aml_checks.created_at"
		params.IsAscOrdering = false
	}

	sb.OrderBy(params.OrderBy)
	if !params.IsAscOrdering {
		sb.Desc()
	}
	sb.Limit(int(limit))
	sb.Offset(int(offset))

	items := make([]*FindRow, 0)
	sql, args := sb.Build()
	if err := pgxscan.Select(ctx, s.psql, &items, sql, args...); err != nil {
		return nil, fmt.Errorf("select aml_checks: %w", err)
	}

	if len(items) > 0 {
		amlCheckIDs := make([]interface{}, len(items))
		for i, item := range items {
			amlCheckIDs[i] = item.ID
		}

		historySb := sqlbuilder.PostgreSQL.NewSelectBuilder()
		historySb.Select(
			`aml_check_history.*`,
		).
			From("aml_check_history").
			Where(historySb.In("aml_check_id", amlCheckIDs...)).
			OrderBy("created_at")

		var historyItems []*models.AmlCheckHistory
		historySQL, historyArgs := historySb.Build()
		if err := pgxscan.Select(ctx, s.psql, &historyItems, historySQL, historyArgs...); err != nil {
			return nil, fmt.Errorf("select aml_check_history: %w", err)
		}

		historyMap := make(map[uuid.UUID][]*models.AmlCheckHistory)
		for _, h := range historyItems {
			historyMap[h.AmlCheckID] = append(historyMap[h.AmlCheckID], h)
		}

		for _, item := range items {
			item.History = historyMap[item.ID]
		}
	}

	var totalCnt uint64
	pagingSQL, args := countSb.Build()
	if err := pgxscan.Get(ctx, s.psql, &totalCnt, pagingSQL, args...); err != nil {
		return nil, fmt.Errorf("select paging query: %w", err)
	}

	var page uint64 = 1
	if params.Page != nil {
		page = uint64(*params.Page)
	}

	var pagesCnt uint64 = 1
	if params.PageSize != nil {
		pagesCnt = uint64(math.Ceil(float64(totalCnt) / float64(*params.PageSize)))
	}

	return &storecmn.FindResponseWithFullPagination[*FindRow]{
		Items: items,
		Pagination: storecmn.FullPagingData{
			Total:    totalCnt,
			PageSize: uint64(limit),
			Page:     page,
			LastPage: pagesCnt,
		},
	}, nil
}
