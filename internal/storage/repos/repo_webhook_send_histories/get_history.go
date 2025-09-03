package repo_webhook_send_histories

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/pkg/dbutils"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
)

type FindRow struct {
	models.WebhookSendHistory
}

type GetHistoriesParams struct {
	storecmn.CommonFindParams
	StoreIDs []uuid.UUID
}

func (s *CustomQuerier) GetByStores(
	ctx context.Context,
	params GetHistoriesParams,
) (*storecmn.FindResponseWithPagingFlag[*FindRow], error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()

	storeIDs := make([]interface{}, len(params.StoreIDs))
	for i, id := range params.StoreIDs {
		storeIDs[i] = id
	}

	sb.Select("webhook_send_histories.*").
		From("webhook_send_histories").
		Where(sb.In("webhook_send_histories.store_id", storeIDs...))

	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize, dbutils.WithMaxLimit(500))
	if err != nil {
		return nil, err
	}

	if params.OrderBy == "" {
		params.OrderBy = "webhook_send_histories.created_at"
	}
	sb.OrderBy(params.OrderBy)
	if params.IsAscOrdering {
		sb.Asc()
	} else {
		sb.Desc()
	}
	increasedLimit := int(limit) + 1
	sb.Limit(increasedLimit)
	sb.Offset(int(offset))

	var items []*FindRow
	sql, args := sb.Build()
	if err := pgxscan.Select(ctx, s.psql, &items, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	var nextPageExists bool
	if len(items) >= increasedLimit {
		nextPageExists = true
		items = items[:limit]
	}

	return &storecmn.FindResponseWithPagingFlag[*FindRow]{
		Items:            items,
		IsNextPageExists: nextPageExists,
	}, nil
}
