package repo_user_exchanges

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ICustomQuerier interface {
	Querier
	ToggleWithdrawals(ctx context.Context, params ToggleWithdrawalParams) (*models.UserExchange, error)
	ChangeSwapState(ctx context.Context, params ChangeSwapStateParams) (*models.UserExchange, error)
	ChangeWithdrawalState(ctx context.Context, params ChangeWithdrawalStateParams) (*models.UserExchange, error)
	DisableAllPerUserExceptExchange(ctx context.Context, userID uuid.UUID, exchangeID uuid.UUID) error
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
