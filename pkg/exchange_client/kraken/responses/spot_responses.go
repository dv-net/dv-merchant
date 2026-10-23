//nolint:tagliatelle
package responses

import "github.com/shopspring/decimal"

const (
	OrderStatusPending  = "pending"
	OrderStatusOpen     = "open"
	OrderStatusClosed   = "closed"
	OrderStatusCanceled = "canceled"
	OrderStatusExpired  = "expired"
)

type AddOrderDescr struct {
	Order string `json:"order"` // "buy 1.25000000 XBTUSD @ market"
	Close string `json:"close,omitempty"`
}

type AddOrderResponse struct {
	Descr AddOrderDescr `json:"descr"`
	TxID  []string      `json:"txid"` // empty when the request had validate=true
}

type OrderDescr struct {
	Pair      string `json:"pair"`
	Type      string `json:"type"`
	OrderType string `json:"ordertype"`
	Price     string `json:"price"`
	Price2    string `json:"price2"`
	Leverage  string `json:"leverage"`
	Order     string `json:"order"`
	Close     string `json:"close"`
}

type Order struct {
	RefID      *string         `json:"refid"`
	UserRef    *int64          `json:"userref"`
	ClOrdID    string          `json:"cl_ord_id"`
	Status     string          `json:"status"`
	Reason     *string         `json:"reason"`
	OpenTm     float64         `json:"opentm"`
	CloseTm    float64         `json:"closetm"`
	StartTm    float64         `json:"starttm"`
	ExpireTm   float64         `json:"expiretm"`
	Descr      OrderDescr      `json:"descr"`
	Vol        decimal.Decimal `json:"vol"`
	VolExec    decimal.Decimal `json:"vol_exec"` // executed volume in base currency
	Cost       decimal.Decimal `json:"cost"`     // executed cost in quote currency
	Fee        decimal.Decimal `json:"fee"`      // in quote currency
	Price      decimal.Decimal `json:"price"`    // average execution price
	StopPrice  decimal.Decimal `json:"stopprice"`
	LimitPrice decimal.Decimal `json:"limitprice"`
	Misc       string          `json:"misc"`
	OFlags     string          `json:"oflags"`
	Trades     []string        `json:"trades"`
}

func (o Order) IsFinal() bool {
	switch o.Status {
	case OrderStatusClosed, OrderStatusCanceled, OrderStatusExpired:
		return true
	default:
		return false
	}
}

// IsFilled reports a fully executed order. Market orders may be closed
// with vol_exec slightly below vol, so only status is checked.
func (o Order) IsFilled() bool {
	return o.Status == OrderStatusClosed
}

// QueryOrdersResponse is keyed by order txid.
type QueryOrdersResponse map[string]Order
