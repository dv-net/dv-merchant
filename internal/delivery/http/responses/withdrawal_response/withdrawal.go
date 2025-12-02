package withdrawal_response

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WithdrawalWalletAddressResponse struct {
	ID                 uuid.UUID `json:"id" format:"uuid"`
	WithdrawalWalletID uuid.UUID `json:"withdrawal_wallet_id" format:"uuid"`
	Name               *string   `json:"name"`
	Address            string    `json:"address"`
	CreatedAt          time.Time `json:"created_at" format:"date-time"`
	UpdatedAt          time.Time `json:"updated_at" format:"date-time"`
	DeletedAt          time.Time `json:"deleted_at" format:"date-time"`
} //	@name	WithdrawalWalletWithAddress

type WithdrawalWithAddressResponse struct {
	ID              uuid.UUID                          `json:"id" format:"uuid"`
	Status          models.WithdrawalStatus            `json:"status"`
	NativeBalance   decimal.NullDecimal                `json:"native_balance"`
	USDBalance      decimal.NullDecimal                `json:"usd_balance"`
	Addressees      []*WithdrawalWalletAddressResponse `json:"addressees"`
	Interval        models.WithdrawalInterval          `json:"interval"`
	Currency        models.CurrencyShort               `json:"currency"`
	Rate            decimal.Decimal                    `json:"rate"`
	LowBalanceRules LowBalanceWithdrawalRuleResponse   `json:"low_balance_rules"`
} //	@name	WithdrawalWalletWithAddress

type WithdrawalRulesByCurrencyResponse struct {
	ID              uuid.UUID                         `json:"id"`
	Status          models.WithdrawalStatus           `json:"status"`
	NativeBalance   decimal.NullDecimal               `json:"native_balance"`
	USDBalance      decimal.NullDecimal               `json:"usd_balance"`
	Addressees      []*models.WithdrawalWalletAddress `json:"addressees"`
	Interval        models.WithdrawalInterval         `json:"interval"`
	Currency        models.CurrencyShort              `json:"currency"`
	Rate            decimal.Decimal                   `json:"rate"`
	LowBalanceRules LowBalanceWithdrawalRuleResponse  `json:"low_balance_rules"`
} //	@name	WithdrawalRulesByCurrencyResponse

type LowBalanceWithdrawalRuleResponse struct {
	Mode          models.MultiWithdrawalMode `json:"mode"`
	ManualAddress *string                    `json:"manual_address"`
} //	@name	LowBalanceWithdrawalRuleResponse

type WithdrawalRuleResponse struct{} //	@name	WithdrawalRule
