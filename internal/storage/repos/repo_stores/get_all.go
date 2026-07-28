package repo_stores

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/constants"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/pkg/dbutils"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5/pgtype"
)

type GetAllFilteredParams struct {
	Status []constants.StoreVerificationStatus
	storecmn.PageParams
}

type StoreWithOwnerEmail struct {
	models.Store
	OwnerEmail pgtype.Text `db:"owner_email"`
}

func (s *CustomQuerier) GetStoresWithFilter(
	ctx context.Context,
	params GetAllFilteredParams,
) (*storecmn.FindResponseWithFullPagination[*StoreWithOwnerEmail], error) {
	pendingFirst := fmt.Sprintf("(stores.verification_status = '%s') DESC", constants.StoreVerificationStatusPending)

	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("stores.*", "users.email AS owner_email").
		From("stores").
		JoinWithOption(sqlbuilder.LeftJoin, "users", "users.id = stores.user_id")
	if len(params.Status) > 0 {
		statuses := make([]any, len(params.Status))
		for i, st := range params.Status {
			statuses[i] = st.String()
		}
		sb.Where(sb.In("stores.verification_status", statuses...))
	}
	sb.OrderBy(pendingFirst, "stores.created_at DESC")

	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize, dbutils.WithMaxLimit(100))
	if err != nil {
		return nil, err
	}

	sb.Limit(int(limit))
	sb.Offset(int(offset))

	items := make([]*StoreWithOwnerEmail, 0)
	sql, args := sb.Build()
	if err := pgxscan.Select(ctx, s.psql, &items, sql, args...); err != nil {
		return nil, err
	}

	countSb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	countSb.Select("COUNT(stores.id)").From("stores")
	if len(params.Status) > 0 {
		statuses := make([]any, len(params.Status))
		for i, st := range params.Status {
			statuses[i] = st.String()
		}
		countSb.Where(countSb.In("stores.verification_status", statuses...))
	}

	var totalCnt uint32
	countSQL, countArgs := countSb.Build()
	if err := pgxscan.Get(ctx, s.psql, &totalCnt, countSQL, countArgs...); err != nil {
		return nil, err
	}

	var currPage uint32
	if params.Page != nil {
		currPage = *params.Page
	}

	return &storecmn.FindResponseWithFullPagination[*StoreWithOwnerEmail]{
		Items: items,
		Pagination: storecmn.FullPagingData{
			Total:    uint64(totalCnt),
			PageSize: uint64(limit),
			Page:     uint64(currPage),
			LastPage: uint64((totalCnt + limit - 1) / limit),
		},
	}, nil
}
