package repo_users

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"

	"github.com/dv-net/dv-merchant/pkg/dbutils"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5/pgtype"
)

type GetAllFilteredParams struct {
	Roles []string
	storecmn.CommonFindParams
}

type GetAllFilteredRow struct {
	models.User
	UserRoles pgtype.Array[pgtype.Text] `db:"user_roles" json:"user_roles"`
}

func (s *CustomQuerier) GetAllFiltered(ctx context.Context, params GetAllFilteredParams) (*storecmn.FindResponseWithFullPagination[*GetAllFilteredRow], error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("users.*", "ARRAY_AGG(DISTINCT casbin_rule.v1) AS user_roles").
		From("users").
		JoinWithOption("LEFT", "casbin_rule", "users.id = casbin_rule.v0::uuid").
		Where(sb.Equal("p_type", sqlbuilder.Raw("'g'"))).
		GroupBy("users.id")

	if len(params.Roles) > 0 {
		roles := make([]interface{}, len(params.Roles))
		for i, id := range params.Roles {
			roles[i] = id
		}
		sb.Where(sb.In("casbin_rule.v1", roles...))
	}

	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize, dbutils.WithMaxLimit(100))
	if err != nil {
		return nil, err
	}
	increasedLimit := int(limit) + 1 // #nosec
	sb.Limit(increasedLimit)
	sb.Offset(int(offset)) // #nosec

	var items []*GetAllFilteredRow
	sql, args := sb.Build()
	if err := pgxscan.Select(ctx, s.psql, &items, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	var totalCnt uint32
	if len(items) > 0 {
		totalCnt = uint32(len(items))
	}

	var currPage uint32
	if params.Page != nil {
		currPage = *params.Page
	}

	return &storecmn.FindResponseWithFullPagination[*GetAllFilteredRow]{
		Items: items,
		Pagination: storecmn.FullPagingData{
			Total:    uint64(totalCnt),
			PageSize: uint64(limit),
			Page:     uint64(currPage),
			LastPage: uint64((totalCnt + limit - 1) / limit),
		},
	}, nil
}
