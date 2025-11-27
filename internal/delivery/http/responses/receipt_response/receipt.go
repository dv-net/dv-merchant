package receipt_response

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/shopspring/decimal"
)

type ReceiptResponse struct {
	ID         string               `json:"id" format:"uuid"`
	Status     models.ReceiptStatus `json:"status"`
	StoreID    string               `json:"store_id" format:"uuid"`
	CurrencyID string               `json:"currency_id"`
	Amount     decimal.Decimal      `json:"amount"`
	WalletID   string               `json:"wallet_id"`
	CreatedAt  time.Time            `json:"created_at" format:"date-time"`
	UpdatedAt  time.Time            `json:"updated_at" format:"date-time"`
} //	@name	ReceiptResponse
