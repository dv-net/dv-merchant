package withdrawal_wallet

import (
	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WithdrawalWithAddress struct {
	ID               uuid.UUID                         `json:"id"`
	Status           models.WithdrawalStatus           `json:"status"`
	MinBalanceNative decimal.NullDecimal               `json:"min_balance_native"`
	MinBalanceUsd    decimal.NullDecimal               `json:"min_balance_usd"`
	Addressees       []*models.WithdrawalWalletAddress `json:"addressees"`
	Interval         models.WithdrawalInterval         `json:"interval"`
	Currency         models.CurrencyShort              `json:"currency"`
	Rate             decimal.Decimal                   `json:"rate"`
	MultiWithdrawal  MultiWithdrawalRuleDTO            `json:"multi_withdrawal"`
}

type UpdateRulesDTO struct {
	WithdrawalEnabled       string                  `json:"withdrawal_enabled"`
	WithdrawalMinBalance    decimal.NullDecimal     `json:"withdrawal_min_balance"`
	WithdrawalMinBalanceUsd decimal.NullDecimal     `json:"withdrawal_min_balance_usd"`
	WithdrawalInterval      string                  `json:"withdrawal_interval"`
	Currency                *models.Currency        `json:"currency"`
	UserID                  uuid.UUID               `json:"user_id"`
	MultiWithdrawal         *MultiWithdrawalRuleDTO `json:"multi_withdrawal"`
}

type MultiWithdrawalRuleDTO struct {
	Mode          models.MultiWithdrawalMode `json:"mode"`
	ManualAddress *string                    `json:"manual_address"`
}

type UpdateAddressesListDTO struct {
	WalletID  uuid.UUID
	TOTP      string
	Addresses []AddressDTO
}

type AddressDTO struct {
	Name    *string
	Address string
}
