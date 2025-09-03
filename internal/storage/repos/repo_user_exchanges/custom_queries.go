package repo_user_exchanges

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
)

type ToggleWithdrawalParams struct {
	UserID     uuid.UUID
	ExchangeID uuid.UUID
}

type ToggleWithdrawalStateRow struct {
	models.UserExchange
}

func (s *CustomQuerier) ToggleWithdrawals(ctx context.Context, params ToggleWithdrawalParams) (*models.UserExchange, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("user_exchanges.withdrawal_state").From("user_exchanges").Where(sb.Equal("user_id", params.UserID), sb.Equal("exchange_id", params.ExchangeID))

	var oldState []string
	sql, args := sb.Build()
	if err := pgxscan.Select(ctx, s.psql, &oldState, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	ub := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	ub.Update("user_exchanges").Set(ub.Assign("withdrawal_state", models.ExchangeWithdrawalState(oldState[0]).Invert().String())).
		Where(ub.Equal("user_id", params.UserID), ub.Equal("exchange_id", params.ExchangeID)).SQL("RETURNING *")

	var items []*ToggleWithdrawalStateRow
	sql, args = ub.Build()
	if err := pgxscan.Select(ctx, s.psql, &items, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}
	item := items[0]
	return &models.UserExchange{
		ID:              item.ID,
		ExchangeID:      item.ExchangeID,
		UserID:          item.UserID,
		WithdrawalState: item.WithdrawalState,
		SwapState:       item.SwapState,
	}, nil
}

type ChangeSwapStateParams struct {
	UserID     uuid.UUID
	ExchangeID uuid.UUID
	State      models.ExchangeSwapState
}

type ChangeSwapStateRow struct {
	models.UserExchange
}

func (s *CustomQuerier) ChangeSwapState(ctx context.Context, params ChangeSwapStateParams) (*models.UserExchange, error) { //nolint:dupl
	ub := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	ub.Update("user_exchanges").Set(ub.Assign("swap_state", params.State.String())).
		Where(ub.Equal("user_id", params.UserID), ub.Equal("exchange_id", params.ExchangeID)).SQL("RETURNING *")

	var items []*ChangeSwapStateRow
	sql, args := ub.Build()
	if err := pgxscan.Select(ctx, s.psql, &items, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}
	item := items[0]
	return &models.UserExchange{
		ID:              item.ID,
		ExchangeID:      item.ExchangeID,
		UserID:          item.UserID,
		WithdrawalState: item.WithdrawalState,
		SwapState:       item.SwapState,
	}, nil
}

type ChangeWithdrawalStateParams struct {
	UserID     uuid.UUID
	ExchangeID uuid.UUID
	State      models.ExchangeWithdrawalState
}

type ChangeWithdrawalStateRow struct {
	models.UserExchange
}

func (s *CustomQuerier) ChangeWithdrawalState(ctx context.Context, params ChangeWithdrawalStateParams) (*models.UserExchange, error) { //nolint:dupl
	ub := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	ub.Update("user_exchanges").Set(ub.Assign("withdrawal_state", params.State.String())).
		Where(ub.Equal("user_id", params.UserID), ub.Equal("exchange_id", params.ExchangeID)).SQL("RETURNING *")

	var items []*ChangeWithdrawalStateRow
	sql, args := ub.Build()
	if err := pgxscan.Select(ctx, s.psql, &items, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}
	item := items[0]
	return &models.UserExchange{
		ID:              item.ID,
		ExchangeID:      item.ExchangeID,
		UserID:          item.UserID,
		WithdrawalState: item.WithdrawalState,
		SwapState:       item.SwapState,
	}, nil
}

func (s *CustomQuerier) DisableAllPerUserExceptExchange(ctx context.Context, userID uuid.UUID, exchangeID uuid.UUID) error {
	ub := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	ub.Update("user_exchanges").
		Set(ub.Assign("swap_state", models.ExchangeSwapStateDisabled.String()), ub.Assign("withdrawal_state", models.ExchangeWithdrawalStateDisabled.String())).
		Where(ub.Equal("user_id", userID), ub.NotEqual("exchange_id", exchangeID))

	sql, args := ub.Build()
	if _, err := s.psql.Exec(ctx, sql, args...); err != nil {
		return fmt.Errorf("update: %w", err)
	}
	return nil
}
