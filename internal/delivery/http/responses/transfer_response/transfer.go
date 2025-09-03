package transfer_response

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/shopspring/decimal"
)

type GetPrefetchDataResponse struct {
	Currency      models.CurrencyShort `json:"currency"`
	Amount        decimal.Decimal      `json:"amount"`
	AmountUsd     decimal.Decimal      `json:"amount_usd"`
	Type          models.TransferKind  `json:"type"`
	FromAddresses []string             `json:"from_addresses"`
	ToAddresses   []string             `json:"to_addresses"`
} // @name GetPrefetchDataResponse

type GetTransferResponse struct {
	ID            string                `json:"id" format:"uuid"`
	Number        int64                 `json:"number"`
	Stage         models.TransferStage  `json:"stage"`
	UserID        string                `json:"user_id" format:"uuid"`
	Kind          models.TransferKind   `json:"kind"`
	CurrencyID    string                `json:"currency_id"`
	Status        models.TransferStatus `json:"status"`
	Step          *string               `json:"step"`
	FromAddresses []string              `json:"from_addresses"`
	ToAddresses   []string              `json:"to_addresses"`
	TxHash        *string               `json:"tx_hash"`
	Amount        decimal.Decimal       `json:"amount"`
	AmountUsd     decimal.Decimal       `json:"amount_usd"`
	Message       *string               `json:"message"`
	CreatedAt     *time.Time            `json:"created_at" format:"date-time"`
	UpdatedAt     *time.Time            `json:"updated_at" format:"date-time"`
} // @name GetTransferResponse
