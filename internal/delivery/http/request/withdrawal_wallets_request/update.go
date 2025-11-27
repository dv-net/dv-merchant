package withdrawal_wallets_request

import (
	"errors"

	"github.com/dv-net/dv-merchant/internal/models"

	validation "github.com/go-ozzo/ozzo-validation/v4" // todo replace with validator

	"github.com/dv-net/dv-processing/pkg/avalidator"

	"github.com/shopspring/decimal"
)

type MultiWithdrawalRules struct {
	Mode          models.MultiWithdrawalMode `json:"mode" validate:"required,oneof=random disabled processing manual"`
	ManualAddress *string                    `json:"manual_address" validate:"omitempty,min=1,max=255"`
} //	@name	MultiWithdrawalRules

func (r *MultiWithdrawalRules) ValidateByBlockchain(b models.Blockchain) error {
	if r.Mode == models.MultiWithdrawalModeManual && r.ManualAddress == nil {
		return errors.New("missing required field: manual_address")
	}

	if r.ManualAddress != nil && !avalidator.ValidateAddressByBlockchain(
		*r.ManualAddress,
		b.String(),
	) {
		return errors.New("invalid manual withdraw address")
	}

	return validation.ValidateStruct(r)
}

type UpdateRequest struct {
	Status          models.WithdrawalStatus   `json:"status"`
	MinBalance      decimal.NullDecimal       `json:"min_balance" validate:"omitempty"`
	MinBalanceUSD   decimal.NullDecimal       `json:"min_balance_usd" validate:"omitempty"`
	Interval        models.WithdrawalInterval `json:"interval" validate:"required"`
	LowBalanceRules *MultiWithdrawalRules     `json:"low_balance_rules" validate:"omitempty"`
} //	@name	UpdateWithdrawalWalletsRequest

func (ur *UpdateRequest) Validate() error {
	if ur.MinBalance.Valid && ur.MinBalanceUSD.Valid {
		return errors.New("min_balance and min_balance_usd cannot be set at the same time")
	}

	if ur.MinBalanceUSD.Valid && !ur.MinBalanceUSD.Decimal.GreaterThan(decimal.Zero) {
		return errors.New("min_balance_usd must be positive non-zero")
	}

	if ur.MinBalance.Valid && !ur.MinBalance.Decimal.GreaterThan(decimal.Zero) {
		return errors.New("min_balance must be positive non-zero")
	}

	if !ur.Interval.IsValid() {
		return errors.New("invalid withdrawal_interval")
	}

	return validation.ValidateStruct(ur)
}
