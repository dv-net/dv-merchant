package responses

import "github.com/shopspring/decimal"

const (
	OrderStatusNew               = "NEW"
	OrderStatusFilled            = "FILLED"
	OrderStatusPartiallyFilled   = "PARTIALLY_FILLED"
	OrderStatusCanceled          = "CANCELED"
	OrderStatusPartiallyCanceled = "PARTIALLY_CANCELED"
)

type PlaceOrderResponse struct {
	Symbol        string          `json:"symbol"`
	OrderID       string          `json:"orderId"`
	ClientOrderID string          `json:"clientOrderId"`
	Price         decimal.Decimal `json:"price"`
	OrigQty       decimal.Decimal `json:"origQty"`
	Type          string          `json:"type"`
	Side          string          `json:"side"`
	TransactTime  int64           `json:"transactTime"`
}

type QueryOrderResponse struct {
	Symbol              string          `json:"symbol"`
	OrderID             string          `json:"orderId"`
	ClientOrderID       string          `json:"clientOrderId"`
	Price               decimal.Decimal `json:"price"`
	OrigQty             decimal.Decimal `json:"origQty"`
	ExecutedQty         decimal.Decimal `json:"executedQty"`
	CummulativeQuoteQty decimal.Decimal `json:"cummulativeQuoteQty"`
	Status              string          `json:"status"`
	Type                string          `json:"type"`
	Side                string          `json:"side"`
}
