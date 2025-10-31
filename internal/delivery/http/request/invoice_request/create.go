package invoice_request

import "github.com/shopspring/decimal"

// CreateRequest represents the request body for creating an invoice
type CreateRequest struct {
	// Amount in USD
	AmountUSD decimal.Decimal `json:"amount_usd" validate:"required,gt=0" example:"100.50" binding:"required"`
	// Unique order identifier
	OrderID string `json:"order_id" validate:"required" example:"ORDER-123456" binding:"required"`
}
