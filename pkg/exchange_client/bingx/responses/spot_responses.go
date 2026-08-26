package responses

import "github.com/shopspring/decimal"

const (
	OrderStatusNew             = "NEW"
	OrderStatusPending         = "PENDING"
	OrderStatusPartiallyFilled = "PARTIALLY_FILLED"
	OrderStatusFilled          = "FILLED"
	OrderStatusCanceled        = "CANCELED"
	OrderStatusFailed          = "FAILED"
)

type PlaceOrderResponse struct {
	Symbol        string          `json:"symbol"`
	OrderID       int64           `json:"orderId"`
	ClientOrderID string          `json:"clientOrderID"`
	Price         decimal.Decimal `json:"price"`
	OrigQty       decimal.Decimal `json:"origQty"`
	ExecutedQty   decimal.Decimal `json:"executedQty"`
	Type          string          `json:"type"`
	Side          string          `json:"side"`
	Status        string          `json:"status"`
	TransactTime  int64           `json:"transactTime"`
}

type QueryOrderResponse struct {
	Symbol              string          `json:"symbol"`
	OrderID             int64           `json:"orderId"`
	ClientOrderID       string          `json:"clientOrderID"`
	Price               decimal.Decimal `json:"price"`
	OrigQty             decimal.Decimal `json:"origQty"`
	ExecutedQty         decimal.Decimal `json:"executedQty"`
	CummulativeQuoteQty decimal.Decimal `json:"cummulativeQuoteQty"`
	Status              string          `json:"status"`
	Type                string          `json:"type"`
	Side                string          `json:"side"`
}
