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
)

type GetByUserParams struct {
	storecmn.CommonFindParams
	ServiceSlug *models.AMLSlug
	DateFrom    *time.Time
	DateTo      *time.Time
}

type AmlCheckDTO struct {
	models.AmlCheck
	Slug models.AMLSlug `json:"slug" db:"slug"`
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
	).
		From("aml_checks").
		JoinWithOption("INNER", "aml_services", "aml_services.id = aml_checks.service_id").
		Where(sb.Equal("aml_checks.user_id", usr.ID.String()))

	countSb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	countSb.Select("COUNT(aml_checks.id)").
		From("aml_checks").
		JoinWithOption("INNER", "aml_services", "aml_services.id = aml_checks.service_id").
		Where(countSb.Equal("aml_checks.user_id", usr.ID.String()))

	if params.ServiceSlug != nil {
		sb.Where(sb.Equal("aml_services.slug", string(*params.ServiceSlug)))
		countSb.Where(countSb.Equal("aml_services.slug", string(*params.ServiceSlug)))
	}

	timezone := usr.Location
	if timezone == "" {
		timezone = time.UTC.String()
	}

	if params.DateFrom != nil {
		from := params.DateFrom.UTC().Format(time.RFC3339)
		sb.Where(sb.GreaterEqualThan(fmt.Sprintf("aml_checks.created_at AT TIME ZONE '%s'", timezone), from))
		countSb.Where(countSb.GreaterEqualThan("aml_checks.created_at", from))
	}

	if params.DateTo != nil {
		to := params.DateTo.UTC().Format(time.RFC3339)
		sb.Where(sb.LessEqualThan(fmt.Sprintf("aml_checks.created_at AT TIME ZONE '%s'", timezone), to))
		countSb.Where(countSb.LessEqualThan("aml_checks.created_at", to))
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
