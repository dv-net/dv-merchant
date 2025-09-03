package repo_exchange_orders

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
)

type GetExchangeOrdersByUserAndExchangeIDParams struct {
	UserID     uuid.UUID
	ExchangeID uuid.NullUUID
	DateFrom   *time.Time
	DateTo     *time.Time
	storecmn.CommonFindParams
}

type GetExportExchangeOrderByUserAndExchangeIDParams struct {
	UserID     uuid.UUID
	ExchangeID uuid.NullUUID
	DateFrom   *time.Time
	DateTo     *time.Time
}

type GetExchangeOrdersByStatus struct {
	Statuses []models.ExchangeOrderStatus
}

type GetExchangeOrdersRow struct {
	models.ExchangeOrder
	Slug string `db:"slug" json:"slug"`
}

func (s *CustomQuerier) GetByStatus(ctx context.Context, params GetExchangeOrdersByStatus) ([]*models.ExchangeOrder, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()

	statuses := make([]interface{}, 0, len(params.Statuses))
	for _, status := range params.Statuses {
		statuses = append(statuses, status.String())
	}
	sb.Select("*").From("exchange_orders").
		Where(sb.In("status", statuses...))

	var orderRows []*GetExchangeOrdersRow
	sql, args := sb.Build()
	if err := pgxscan.Select(ctx, s.psql, &orderRows, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	items := make([]*models.ExchangeOrder, 0, len(orderRows))
	for _, row := range orderRows {
		items = append(items, &models.ExchangeOrder{
			ID:              row.ID,
			ExchangeID:      row.ExchangeID,
			ExchangeOrderID: row.ExchangeOrderID,
			ClientOrderID:   row.ClientOrderID,
			Symbol:          row.Symbol,
			Side:            row.Side,
			Amount:          row.Amount,
			AmountUsd:       row.AmountUsd,
			OrderCreatedAt:  row.OrderCreatedAt,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
			FailReason:      row.FailReason,
			Status:          row.Status,
			UserID:          row.UserID,
		})
	}

	return items, nil
}

type GetExchangeOrdersByUserAndExchangeIDRow struct {
	models.ExchangeOrder
	Slug string `db:"slug" json:"slug"`
}

func (s *CustomQuerier) GetByUserAndExchangeID(ctx context.Context, params GetExchangeOrdersByUserAndExchangeIDParams) (*storecmn.FindResponseWithFullPagination[*GetExchangeOrdersByUserAndExchangeIDRow], error) {
	countSb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	countSb.Select("COUNT(exchange_orders.id)").
		From("exchange_orders").
		Where(countSb.Equal("user_id", params.UserID))

	if params.ExchangeID.Valid {
		countSb.Where(countSb.Equal("exchange_id", params.ExchangeID))
	}

	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("exchange_orders.*, exchanges.slug AS slug").From("exchange_orders").
		JoinWithOption(sqlbuilder.LeftJoin, "exchanges ON exchange_orders.exchange_id = exchanges.id").
		Where(sb.Equal("user_id", params.UserID))

	if params.ExchangeID.Valid {
		sb.Where(sb.Equal("exchange_id", params.ExchangeID))
	}

	if params.DateFrom != nil {
		sb.Where(sb.GreaterEqualThan("exchange_orders.created_at", params.DateFrom))
		countSb.Where(countSb.GreaterEqualThan("exchange_orders.created_at", params.DateFrom))
	}

	if params.DateTo != nil {
		sb.Where(sb.LessEqualThan("exchange_orders.created_at", params.DateTo))
		countSb.Where(countSb.LessEqualThan("exchange_orders.created_at", params.DateTo))
	}

	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize, dbutils.WithMaxLimit(500))
	if err != nil {
		return nil, err
	}

	if params.OrderBy == "" {
		params.OrderBy = "exchange_orders.created_at"
	}
	sb.OrderBy(params.OrderBy)

	if !params.IsAscOrdering {
		sb.Desc()
	}

	sb.Limit(int(limit))   // #nosec
	sb.Offset(int(offset)) // #nosec

	var orderRows []*GetExchangeOrdersRow
	sql, args := sb.Build()
	if err := pgxscan.Select(ctx, s.psql, &orderRows, sql, args...); err != nil {
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

	items := make([]*GetExchangeOrdersByUserAndExchangeIDRow, 0, len(orderRows))
	for _, row := range orderRows {
		items = append(items, &GetExchangeOrdersByUserAndExchangeIDRow{
			ExchangeOrder: models.ExchangeOrder{
				ID:              row.ID,
				ExchangeID:      row.ExchangeID,
				ExchangeOrderID: row.ExchangeOrderID,
				ClientOrderID:   row.ClientOrderID,
				Symbol:          row.Symbol,
				Side:            row.Side,
				Amount:          row.Amount,
				AmountUsd:       row.AmountUsd,
				OrderCreatedAt:  row.OrderCreatedAt,
				CreatedAt:       row.CreatedAt,
				UpdatedAt:       row.UpdatedAt,
				FailReason:      row.FailReason,
				Status:          row.Status,
				UserID:          row.UserID,
			},
			Slug: row.Slug,
		})
	}

	return &storecmn.FindResponseWithFullPagination[*GetExchangeOrdersByUserAndExchangeIDRow]{
		Items: items,
		Pagination: storecmn.FullPagingData{
			Total:    uint64(totalCnt),
			PageSize: uint64(limit),
			Page:     uint64(currPage),
			LastPage: pagesCnt,
		},
	}, nil
}

func (s *CustomQuerier) GetAllByUserFiltered(ctx context.Context, params GetExportExchangeOrderByUserAndExchangeIDParams) ([]*GetExchangeOrdersByUserAndExchangeIDRow, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("exchange_orders.*, exchanges.slug AS slug").From("exchange_orders").
		JoinWithOption(sqlbuilder.LeftJoin, "exchanges ON exchange_orders.exchange_id = exchanges.id").
		Where(sb.Equal("user_id", params.UserID))

	if params.ExchangeID.Valid {
		sb.Where(sb.Equal("exchange_id", params.ExchangeID))
	}

	if params.DateFrom != nil {
		sb.Where(sb.GreaterEqualThan("created_at", params.DateFrom))
	}

	if params.DateTo != nil {
		sb.Where(sb.LessEqualThan("created_at", params.DateTo))
	}

	var orderRows []*GetExchangeOrdersRow
	sql, args := sb.Build()
	if err := pgxscan.Select(ctx, s.psql, &orderRows, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	items := make([]*GetExchangeOrdersByUserAndExchangeIDRow, 0, len(orderRows))
	for _, row := range orderRows {
		items = append(items, &GetExchangeOrdersByUserAndExchangeIDRow{
			ExchangeOrder: models.ExchangeOrder{
				ID:              row.ID,
				ExchangeID:      row.ExchangeID,
				ExchangeOrderID: row.ExchangeOrderID,
				ClientOrderID:   row.ClientOrderID,
				Symbol:          row.Symbol,
				Side:            row.Side,
				Amount:          row.Amount,
				AmountUsd:       row.AmountUsd,
				OrderCreatedAt:  row.OrderCreatedAt,
				CreatedAt:       row.CreatedAt,
				UpdatedAt:       row.UpdatedAt,
				FailReason:      row.FailReason,
				Status:          row.Status,
				UserID:          row.UserID,
			},
			Slug: row.Slug,
		})
	}
	return items, nil
}
