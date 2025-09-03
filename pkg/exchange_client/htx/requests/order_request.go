//nolint:tagliatelle
package requests

import (
	"strings"

	htxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/models"
)

type GetOrderHistoryRequest struct {
	Symbol    string `json:"symbol" url:"symbol"` // All supported trading symbols, e.g. btcusdt, bccbtc
	Types     string `json:"types" url:"types,omitempty"`
	StartTime int64  `json:"start-time" url:"start-time,omitempty"` // Search starts time, UTC time in millisecond
	EndTime   int64  `json:"end-time" url:"end-time,omitempty"`     // Search ends time, UTC time in millisecond
	States    string `json:"states" url:"states"`
	From      string `json:"from" url:"from,omitempty"`
	Direct    string `json:"direct" url:"direct,omitempty"` // next, prev. Search direction when 'from' is used
	Size      string `json:"size" url:"size,omitempty"`
}

func (o *GetOrderHistoryRequest) Default() {
	o.Symbol = "btcusdt"
	o.States = strings.Join([]string{
		htxmodels.OrderStateFilled.String(),
		htxmodels.OrderStateCanceled.String(),
		htxmodels.OrderStateSubmitted.String(),
		htxmodels.OrderStateCreated.String(),
		htxmodels.OrderStatePartialFilled.String(),
		htxmodels.OrderStatePartialCanceled.String(),
	}, ",")
}

type PlaceOrderRequest struct {
	AccountID        string `json:"account-id"`
	Symbol           string `json:"symbol"`
	Type             string `json:"type"`
	Amount           string `json:"amount"`
	Price            string `json:"price,omitempty"`
	Source           string `json:"source,omitempty"`
	ClientOrderID    string `json:"client-order-id,omitempty"`
	SelfMatchPrevent int    `json:"self-match-prevent,omitempty"`
	StopPrice        string `json:"stop-price,omitempty"`
	Operator         string `json:"operator,omitempty"`
}

type GetOrderByClientIDRequest struct {
	ClientOrderID string `url:"clientOrderId"`
}
