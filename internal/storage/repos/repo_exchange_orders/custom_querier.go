package repo_exchange_orders

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"

	"github.com/jackc/pgx/v5"
)

type ICustomQuerier interface {
	Querier
	GetByUserAndExchangeID(context.Context, GetExchangeOrdersByUserAndExchangeIDParams) (*storecmn.FindResponseWithFullPagination[*GetExchangeOrdersByUserAndExchangeIDRow], error)
	GetAllByUserFiltered(context.Context, GetExportExchangeOrderByUserAndExchangeIDParams) ([]*GetExchangeOrdersByUserAndExchangeIDRow, error)
	GetByStatus(context.Context, GetExchangeOrdersByStatus) ([]*models.ExchangeOrder, error)
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
