//nolint:tagliatelle
package models

type Order struct {
	OrderID            string      `json:"orderId"`
	OrderLinkID        string      `json:"orderLinkId"`
	BlockTradeID       string      `json:"blockTradeId"`
	Symbol             string      `json:"symbol"`
	Price              string      `json:"price"`
	Qty                string      `json:"qty"`
	CumExecQty         string      `json:"cumExecQty"`
	Side               string      `json:"side"`
	OrderStatus        OrderStatus `json:"orderStatus"`
	CancelType         string      `json:"cancelType"`
	OrderType          string      `json:"orderType"`
	LastPriceOnCreated string      `json:"lastPriceOnCreated"`
	ReduceOnly         bool        `json:"reduceOnly"`
	CreatedTime        string      `json:"createdTime"`
	UpdatedTime        string      `json:"updatedTime"`
	ExtraFees          string      `json:"extraFees"`
}
