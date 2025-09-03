package repo_transfers

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

type GetTransfersByUserAndStatusParams struct {
	storecmn.CommonFindParams
	UserID   uuid.UUID
	Stages   []models.TransferStage
	Kinds    []models.TransferKind
	DateFrom *time.Time
}

type GetTransfersByUserAndStatusRow struct {
	models.Transfer
}

func (s *CustomQuerier) GetTransfersByUserAndStatus(ctx context.Context, params GetTransfersByUserAndStatusParams) (*storecmn.FindResponseWithFullPagination[*GetTransfersByUserAndStatusRow], error) {
	stages := sqlbuilder.Flatten(params.Stages)
	kinds := sqlbuilder.Flatten(params.Kinds)

	dateFrom := time.Now().Add(-time.Hour * 168)
	if params.DateFrom != nil {
		dateFrom = *params.DateFrom
	}

	countSb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	countSb.Select("COUNT(1) AS count").
		From("transfers").
		Where(countSb.Equal("user_id", params.UserID)).
		Where(countSb.In("stage", stages...)).
		Where(countSb.GreaterThan("created_at", dateFrom))

	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()

	sb.Select("*").
		From("transfers").
		Where(sb.Equal("user_id", params.UserID)).
		Where(sb.In("stage", stages...)).
		Where(sb.GreaterThan("created_at", dateFrom)).
		OrderBy("number DESC")

	if len(kinds) > 0 {
		sb.Where(sb.Any("kind", "=", kinds))
		countSb.Where(countSb.Any("kind", "=", kinds))
	}

	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize, dbutils.WithMaxLimit(1000))
	if err != nil {
		return nil, err
	}

	increasedLimit := int(limit) // #nosec
	sb.Limit(increasedLimit)
	sb.Offset(int(offset)) // #nosec

	var transfers []*GetTransfersByUserAndStatusRow
	sql, args := sb.Build()

	if err := pgxscan.Select(ctx, s.psql, &transfers, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	var totalCnt uint64
	pagingSQL, args := countSb.Build()
	if err := pgxscan.Get(ctx, s.psql, &totalCnt, pagingSQL, args...); err != nil {
		return nil, fmt.Errorf("select paging query: %w", err)
	}

	var currPage uint32
	if params.Page != nil {
		currPage = *params.Page
	}

	var pagesCnt uint64 = 1
	if params.PageSize != nil {
		pagesCnt = uint64(math.Ceil(float64(totalCnt) / float64(*params.PageSize)))
	}

	return &storecmn.FindResponseWithFullPagination[*GetTransfersByUserAndStatusRow]{
		Items: transfers,
		Pagination: storecmn.FullPagingData{
			Total:    totalCnt,
			PageSize: uint64(limit),
			Page:     uint64(currPage),
			LastPage: pagesCnt,
		},
	}, nil
}

type GetTransferHistoryByAddressParams struct {
	storecmn.CommonFindParams
	Stages   []models.TransferStage
	Kinds    []models.TransferKind
	UserID   uuid.UUID
	DateFrom *time.Time
	Address  string
}

type GetTransfersHistoryByAddressRow struct {
	models.Transfer
}

func (s *CustomQuerier) GetTransfersHistoryByAddress(ctx context.Context, params GetTransferHistoryByAddressParams) (*storecmn.FindResponseWithFullPagination[*GetTransfersHistoryByAddressRow], error) {
	stages := sqlbuilder.Flatten(params.Stages)
	kinds := sqlbuilder.Flatten(params.Kinds)

	dateFrom := time.Now().Add(-time.Hour * 168)
	if params.DateFrom != nil {
		dateFrom = *params.DateFrom
	}

	countSb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	countSb.Select("COUNT(1) AS count").
		From("transfers").
		Where(countSb.Equal("user_id", params.UserID)).
		Where(countSb.In("stage", stages...)).
		Where(countSb.GreaterThan("created_at", dateFrom)).
		Where(
			countSb.Or(
				countSb.In("from_addresses", sqlbuilder.Flatten([]string{params.Address})),
				countSb.In("to_addresses", sqlbuilder.Flatten([]string{params.Address})),
			))

	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("*").
		From("transfers").
		Where(sb.Equal("user_id", params.UserID)).
		Where(sb.In("stage", stages...)).
		Where(sb.GreaterThan("created_at", dateFrom)).
		Where(
			sb.Or(
				sb.In("from_addresses", sqlbuilder.Flatten([]string{params.Address})),
				sb.In("to_addresses", sqlbuilder.Flatten([]string{params.Address})),
			)).
		OrderBy("number DESC")

	if len(kinds) > 0 {
		sb.Where(sb.Any("kind", "=", kinds))
		countSb.Where(countSb.Any("kind", "=", kinds))
	}

	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize, dbutils.WithMaxLimit(1000))
	if err != nil {
		return nil, err
	}

	increasedLimit := int(limit) // #nosec
	sb.Limit(increasedLimit)
	sb.Offset(int(offset)) // #nosec

	var transfers []*GetTransfersHistoryByAddressRow
	sql, args := sb.Build()

	if err := pgxscan.Select(ctx, s.psql, &transfers, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	var totalCnt uint64
	pagingSQL, args := countSb.Build()
	if err := pgxscan.Get(ctx, s.psql, &totalCnt, pagingSQL, args...); err != nil {
		return nil, fmt.Errorf("select paging query: %w", err)
	}

	var currPage uint32
	if params.Page != nil {
		currPage = *params.Page
	}

	var pagesCnt uint64 = 1
	if params.PageSize != nil {
		pagesCnt = uint64(math.Ceil(float64(totalCnt) / float64(*params.PageSize)))
	}

	return &storecmn.FindResponseWithFullPagination[*GetTransfersHistoryByAddressRow]{
		Items: transfers,
		Pagination: storecmn.FullPagingData{
			Total:    totalCnt,
			PageSize: uint64(limit),
			Page:     uint64(currPage),
			LastPage: pagesCnt,
		},
	}, nil
}
