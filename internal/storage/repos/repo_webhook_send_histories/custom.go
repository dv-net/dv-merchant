package repo_webhook_send_histories

import (
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"

	"github.com/jackc/pgx/v5"
	"golang.org/x/net/context"
)

type ICustomQuerier interface {
	Querier
	GetByStores(ctx context.Context, params GetHistoriesParams) (*storecmn.FindResponseWithPagingFlag[*FindRow], error)
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
