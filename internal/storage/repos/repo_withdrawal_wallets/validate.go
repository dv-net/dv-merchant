package repo_withdrawal_wallets

import (
	"errors"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
)

func (s UpdateParams) Validate() error {
	if s.WithdrawalMinBalance.Valid && s.WithdrawalMinBalance.Decimal.IsNegative() {
		return errors.New("withdrawal_min_balance cannot be negative")
	}

	if !s.WithdrawalMinBalance.Decimal.IsZero() && s.WithdrawalMinBalanceUsd.Valid {
		return errors.New("withdrawal_min_balance cannot be zero or negative")
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
