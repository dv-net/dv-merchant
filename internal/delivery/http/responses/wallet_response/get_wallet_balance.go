package wallet_response

import (
	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"

	"github.com/shopspring/decimal"
)

type GetWalletBalanceResponse struct {
	ID         uuid.UUID         `json:"id"`
	CurrencyID string            `json:"currency_id"`
	Address    string            `json:"address"`
	Blockchain models.Blockchain `json:"blockchain"`
	Amount     decimal.Decimal   `json:"amount"`
	AmountUSD  decimal.Decimal   `json:"amount_usd"`
	Dirty      bool              `json:"dirty"`
} // @name GetWalletBalanceResponse
