package repo_withdrawal_wallets

import (
	"errors"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
)

func (s UpdateParams) Validate() error {
	if s.WithdrawalMinBalance.Valid && !s.WithdrawalMinBalance.Decimal.IsPositive() {
		return errors.New("withdrawal_min_balance must be greater than 0")
	}

	if s.WithdrawalMinBalanceUsd.Valid && !s.WithdrawalMinBalanceUsd.Decimal.IsPositive() {
		return errors.New("withdrawal_min_balance_usd must be greater than 0")
	}

	if s.WithdrawalMinBalance.Valid && s.WithdrawalMinBalanceUsd.Valid {
		return errors.New("withdrawal_min_balance and withdrawal_min_balance_usd cannot be set at the same time")
	}

	if !models.WithdrawalInterval(s.WithdrawalInterval).IsValid() {
		return errors.New("invalid withdrawal_interval")
	}

	if s.CurrencyID == "" {
		return errors.New("currency_id cannot be empty")
	}

	if s.UserID == uuid.Nil {
		return errors.New("user_id cannot be zero UUID")
	}

	return nil
}
