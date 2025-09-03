package repo_exchange_withdrawal_history

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/pkg/dbutils"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
)

type GetWithdrawalHistoryByUserAndExchangeIDParams struct {
	UserID     uuid.UUID
	ExchangeID uuid.NullUUID
	Currency   string
	DateFrom   *time.Time
	DateTo     *time.Time
	storecmn.CommonFindParams
}

type GetWithdrawalExportHistoryByUserAndExchangeIDParams struct {
	UserID     uuid.UUID
	ExchangeID uuid.NullUUID
	Currency   string
	DateFrom   *time.Time
	DateTo     *time.Time
}

type GetLastWithdrawalHistoryRow struct {
	models.ExchangeWithdrawalHistory
}

type GetWithdrawalHistoryByUserAndExchangeIDRow struct {
	ID                     uuid.UUID                      `db:"id" json:"id"`
	UserID                 uuid.UUID                      `db:"user_id" json:"user_id"`
	ExchangeID             uuid.UUID                      `db:"exchange_id" json:"exchange_id"`
	ExchangeOrderID        pgtype.Text                    `db:"exchange_order_id" json:"exchange_order_id"`
	Address                string                         `db:"address" json:"address"`
	MinAmount              decimal.Decimal                `db:"min_amount" json:"min_amount"`
	NativeAmount           decimal.NullDecimal            `db:"native_amount" json:"native_amount"`
	FiatAmount             decimal.NullDecimal            `db:"fiat_amount" json:"fiat_amount"`
	Currency               string                         `db:"currency" json:"currency"`
	Chain                  string                         `db:"chain" json:"chain"`
	Status                 models.WithdrawalHistoryStatus `db:"status" json:"status"`
	Txid                   pgtype.Text                    `db:"txid" json:"txid"`
	CreatedAt              pgtype.Timestamp               `db:"created_at" json:"created_at"`
	UpdatedAt              pgtype.Timestamp               `db:"updated_at" json:"updated_at"`
	Slug                   models.ExchangeSlug            `db:"slug" json:"slug"`
	FailReason             pgtype.Text                    `db:"fail_reason" json:"fail_reason"`
	ExchangeConnectionHash pgtype.Text                    `db:"exchange_connection_hash" json:"exchange_connection_hash"`
}

func (s *CustomQuerier) GetAllByUser(ctx context.Context, params GetWithdrawalExportHistoryByUserAndExchangeIDParams) ([]*models.ExchangeWithdrawalHistoryDTO, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("ewh.*, e.slug").From(sb.As("exchange_withdrawal_history", "ewh")).
		JoinWithOption(sqlbuilder.LeftJoin, sb.As("exchanges", "e"), "ewh.exchange_id = e.id").
		Where(sb.Equal("user_id", params.UserID))

	if params.ExchangeID.Valid {
		sb.Where(sb.Equal("exchange_id", params.ExchangeID))
	}

	if len(params.Currency) > 0 {
		sb.Where(sb.Equal("currency", params.Currency))
	}

	if params.DateFrom != nil {
		sb.Where(sb.GreaterEqualThan("ewh.created_at", params.DateFrom))
	}

	if params.DateTo != nil {
		sb.Where(sb.LessEqualThan("ewh.created_at", params.DateTo))
	}

	var historyRows []*GetWithdrawalHistoryByUserAndExchangeIDRow
	sql, args := sb.Build()
	if err := pgxscan.Select(ctx, s.psql, &historyRows, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	items := make([]*models.ExchangeWithdrawalHistoryDTO, 0, len(historyRows))
	for _, row := range historyRows {
		items = append(items, &models.ExchangeWithdrawalHistoryDTO{
			ID:              row.ID,
			UserID:          row.UserID,
			ExchangeID:      row.ExchangeID,
			ExchangeOrderID: row.ExchangeOrderID,
			Address:         row.Address,
			MinAmount:       row.MinAmount,
			NativeAmount:    row.NativeAmount,
			FiatAmount:      row.FiatAmount,
			Currency:        row.Currency,
			Chain:           row.Chain,
			Status:          row.Status,
			Txid:            row.Txid,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
			Slug:            row.Slug,
			FailReason:      row.FailReason,
		})
	}
	return items, nil
}

func (s *CustomQuerier) GetAllByUserAndExchangeID(ctx context.Context, params GetWithdrawalHistoryByUserAndExchangeIDParams) (*storecmn.FindResponseWithFullPagination[*models.ExchangeWithdrawalHistoryDTO], error) {
	countSb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	countSb.Select("COUNT(exchange_withdrawal_history.id)").
		From("exchange_withdrawal_history").
		Where(countSb.Equal("user_id", params.UserID))

	if params.ExchangeID.Valid {
		countSb.Where(countSb.Equal("exchange_id", params.ExchangeID))
	}

	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("ewh.*, e.slug").From(sb.As("exchange_withdrawal_history", "ewh")).
		JoinWithOption(sqlbuilder.LeftJoin, sb.As("exchanges", "e"), "ewh.exchange_id = e.id").
		Where(sb.Equal("user_id", params.UserID))

	if params.ExchangeID.Valid {
		sb.Where(sb.Equal("exchange_id", params.ExchangeID))
	}

	if len(params.Currency) > 0 {
		sb.Where(sb.Equal("currency", params.Currency))
		countSb.Where(countSb.Equal("currency", params.Currency))
	}

	if params.DateFrom != nil {
		sb.Where(sb.GreaterEqualThan("ewh.created_at", params.DateFrom))
		countSb.Where(countSb.GreaterEqualThan("created_at", params.DateFrom))
	}

	if params.DateTo != nil {
		sb.Where(sb.LessEqualThan("ewh.created_at", params.DateTo))
		countSb.Where(countSb.LessEqualThan("created_at", params.DateTo))
	}

	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize, dbutils.WithMaxLimit(500))
	if err != nil {
		return nil, err
	}

	if params.OrderBy == "" {
		params.OrderBy = "ewh.created_at"
	}
	sb.OrderBy(params.OrderBy)

	if !params.IsAscOrdering {
		sb.Desc()
	}

	sb.Limit(int(limit))   // #nosec
	sb.Offset(int(offset)) // #nosec

	var historyRows []*GetWithdrawalHistoryByUserAndExchangeIDRow
	sql, args := sb.Build()
	if err := pgxscan.Select(ctx, s.psql, &historyRows, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	var totalCnt uint32
	pagingSQL, args := countSb.Build()
	if err := pgxscan.Get(ctx, s.psql, &totalCnt, pagingSQL, args...); err != nil {
		return nil, fmt.Errorf("select count: %w", err)
	}

	var currPage uint32 = 1
	if params.Page != nil {
		currPage = *params.Page
	}

	var pagesCnt uint64 = 1
	if params.PageSize != nil {
		pagesCnt = uint64(math.Ceil(float64(totalCnt) / float64(*params.PageSize)))
	}

	items := make([]*models.ExchangeWithdrawalHistoryDTO, 0, len(historyRows))
	for _, row := range historyRows {
		items = append(items, &models.ExchangeWithdrawalHistoryDTO{
			ID:              row.ID,
			UserID:          row.UserID,
			ExchangeID:      row.ExchangeID,
			ExchangeOrderID: row.ExchangeOrderID,
			Address:         row.Address,
			MinAmount:       row.MinAmount,
			NativeAmount:    row.NativeAmount,
			FiatAmount:      row.FiatAmount,
			Currency:        row.Currency,
			Chain:           row.Chain,
			Status:          row.Status,
			Txid:            row.Txid,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
			Slug:            row.Slug,
			FailReason:      row.FailReason,
		})
	}

	return &storecmn.FindResponseWithFullPagination[*models.ExchangeWithdrawalHistoryDTO]{
		Items: items,
		Pagination: storecmn.FullPagingData{
			Total:    uint64(totalCnt),
			PageSize: uint64(limit),
			Page:     uint64(currPage),
			LastPage: pagesCnt,
		},
	}, nil
}

var unprocessedStatuses = []interface{}{
	models.WithdrawalHistoryStatusNew,
	models.WithdrawalHistoryStatusInProgress,
}

func (s *CustomQuerier) GetLast(ctx context.Context, userID, exchangeID uuid.UUID) (*models.ExchangeWithdrawalHistory, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("*").
		From("exchange_withdrawal_history").
		Where(
			sb.Equal("user_id", userID),
			sb.Equal("exchange_id", exchangeID),
			sb.In("status", unprocessedStatuses...),
		).
		OrderBy("created_at").
		Limit(1)

	var historyRow GetLastWithdrawalHistoryRow
	sql, args := sb.Build()
	if err := pgxscan.Get(ctx, s.psql, &historyRow, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	return &historyRow.ExchangeWithdrawalHistory, nil
}

func (s *CustomQuerier) GetAllUnprocessed(ctx context.Context) ([]*models.ExchangeWithdrawalHistory, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	st := make([]interface{}, 0, 2)
	st = append(st, unprocessedStatuses...)
	sb.Select("*").
		From("exchange_withdrawal_history").
		Where(sb.In("status", st...)).
		Limit(100)

	var historyRows []*models.ExchangeWithdrawalHistory
	sql, args := sb.Build()
	if err := pgxscan.Select(ctx, s.psql, &historyRows, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	return historyRows, nil
}
