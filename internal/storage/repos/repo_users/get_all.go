package repo_users

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/pkg/dbutils"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5/pgtype"
)

type GetAllFilteredParams struct {
	Roles          []string
	ExcludeUserIDs []uuid.UUID
	storecmn.CommonFindParams
}

type GetAllFilteredRow struct {
	models.User
	UserRoles pgtype.Array[pgtype.Text] `db:"user_roles" json:"user_roles"`
}

func (s *CustomQuerier) GetAllFiltered(ctx context.Context, params GetAllFilteredParams) (*storecmn.FindResponseWithFullPagination[*GetAllFilteredRow], error) {
	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize, dbutils.WithMaxLimit(100))
	if err != nil {
		return nil, err
	}

	roles := make([]any, len(params.Roles))
	for i, role := range params.Roles {
		roles[i] = role
	}

	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("users.*", "ARRAY_AGG(DISTINCT casbin_rule.v1) AS user_roles").
		From("users")
	applyUsersListJoinsAndFilters(sb, roles, params.ExcludeUserIDs)
	sb.GroupBy("users.id")
	sb.OrderBy("users.created_at DESC")
	sb.Limit(int(limit))   // #nosec G115
	sb.Offset(int(offset)) // #nosec G115

	var items []*GetAllFilteredRow
	sql, args := sb.Build()
	if err := pgxscan.Select(ctx, s.psql, &items, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	countSb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	countSb.Select("COUNT(DISTINCT users.id)").From("users")
	applyUsersListJoinsAndFilters(countSb, roles, params.ExcludeUserIDs)

	var totalCnt uint64
	countSQL, countArgs := countSb.Build()
	if err := pgxscan.Get(ctx, s.psql, &totalCnt, countSQL, countArgs...); err != nil {
		return nil, fmt.Errorf("count: %w", err)
	}

	var currPage uint32
	if params.Page != nil {
		currPage = *params.Page
	}

	var lastPage uint64
	if limit > 0 {
		lastPage = (totalCnt + uint64(limit) - 1) / uint64(limit)
	}

	return &storecmn.FindResponseWithFullPagination[*GetAllFilteredRow]{
		Items: items,
		Pagination: storecmn.FullPagingData{
			Total:    totalCnt,
			PageSize: uint64(limit),
			Page:     uint64(currPage),
			LastPage: lastPage,
		},
	}, nil
}

func applyUsersListJoinsAndFilters(sb *sqlbuilder.SelectBuilder, roles []any, excludeUserIDs []uuid.UUID) {
	sb.JoinWithOption(
		sqlbuilder.InnerJoin,
		"casbin_rule",
		"users.id = casbin_rule.v0::uuid AND casbin_rule.p_type = 'g'",
	)
	if len(roles) > 0 {
		sb.Where(sb.In("casbin_rule.v1", roles...))
	}
	if len(excludeUserIDs) > 0 {
		excluded := make([]any, len(excludeUserIDs))
		for i, id := range excludeUserIDs {
			excluded[i] = id
		}
		sb.Where(sb.NotIn("users.id", excluded...))
	}
}
