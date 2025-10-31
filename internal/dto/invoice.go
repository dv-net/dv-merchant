package dto

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/shopspring/decimal"
)

type CreateInvoiceDTO struct {
	AmountUSD decimal.Decimal `json:"amount"`
	OrderID   string          `json:"order_id"`
	User      *models.User    `json:"user"`
	Store     *models.Store   `json:"store"`
}

type InvoiceAddressDTO struct {
	Address           string              `json:"address"`
	CurrencyID        string              `json:"currency_id"`
	Blockchain        models.Blockchain   `json:"blockchain"`
	RateAtCreation    decimal.NullDecimal `json:"rate_at_creation"`
	ExpectedAmount    decimal.Decimal     `json:"expected_amount"`
	ExpectedAmountUsd decimal.Decimal     `json:"expected_amount_usd"`
	ReceivedAmount    decimal.Decimal     `json:"received_amount"`
	ReceivedAmountUsd decimal.Decimal     `json:"received_amount_usd"`
	TxHash            *string             `json:"tx_hash"`
	BcUniqKey         *string             `json:"bc_uniq_key"`
	PaidAt            time.Time           `json:"paid_at"`
}

type InvoiceInfoDTO struct {
	Invoice        *models.Invoice      `json:"invoice"`
	InvoiceAddress []*InvoiceAddressDTO `json:"invoice_address"`
}
