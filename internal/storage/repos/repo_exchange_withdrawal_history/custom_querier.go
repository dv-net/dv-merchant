package repo_exchange_withdrawal_history

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ICustomQuerier interface {
	Querier
	GetAllByUserAndExchangeID(ctx context.Context, params GetWithdrawalHistoryByUserAndExchangeIDParams) (*storecmn.FindResponseWithFullPagination[*models.ExchangeWithdrawalHistoryDTO], error)
	GetAllByUser(ctx context.Context, params GetWithdrawalExportHistoryByUserAndExchangeIDParams) ([]*models.ExchangeWithdrawalHistoryDTO, error)
	GetLast(ctx context.Context, userID, exchangeID uuid.UUID) (*models.ExchangeWithdrawalHistory, error)
	GetAllUnprocessed(ctx context.Context) ([]*models.ExchangeWithdrawalHistory, error)
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
