package repo_transfers

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
)

type ICustomQuerier interface {
	Querier
	GetTransfersByUserAndStatus(ctx context.Context, params GetTransfersByUserAndStatusParams) (*storecmn.FindResponseWithFullPagination[*GetTransfersByUserAndStatusRow], error)
	GetTransfersHistoryByAddress(ctx context.Context, params GetTransferHistoryByAddressParams) (*storecmn.FindResponseWithFullPagination[*GetTransfersHistoryByAddressRow], error)
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
