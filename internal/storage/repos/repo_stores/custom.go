package repo_stores

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/jackc/pgx/v5"
)

type ICustomQuerier interface {
	Querier
	GetStoresWithFilter(ctx context.Context, params GetAllFilteredParams) (*storecmn.FindResponseWithFullPagination[*StoreWithOwnerEmail], error)
}

type CustomQuerier struct {
	*Queries
	psql DBTX
}

func NewCustom(psql DBTX) *CustomQuerier {
	return &CustomQuerier{
		Queries: New(psql),
		psql:    psql,
	}
}

func (s *CustomQuerier) WithTx(tx pgx.Tx) *CustomQuerier {
	return &CustomQuerier{
		Queries: New(tx),
		psql:    tx,
	}
}
