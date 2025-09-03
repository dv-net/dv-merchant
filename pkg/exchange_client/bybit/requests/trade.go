//nolint:tagliatelle
package requests

type GetActiveOrdersRequest struct {
	Category    string `json:"category" url:"category"`
	Symbol      string `json:"symbol,omitempty" url:"symbol,omitempty"`
	OrderID     string `json:"orderId,omitempty" url:"orderId,omitempty"`
	OrderLinkID string `json:"orderLinkId,omitempty" url:"orderLinkId,omitempty"`
}

type GetOrderHistoryRequest struct {
	Category    string `json:"category" url:"category"`
	Symbol      string `json:"symbol,omitempty" url:"symbol,omitempty"`
	OrderID     string `json:"orderId,omitempty" url:"orderId,omitempty"`
	OrderLinkID string `json:"orderLinkId,omitempty" url:"orderLinkId,omitempty"`
}

type PlaceOrderRequest struct {
	Category    string `json:"category" url:"category"`
	Symbol      string `json:"symbol" url:"symbol"`
	Side        string `json:"side" url:"side"`
	OrderType   string `json:"orderType" url:"orderType"`
	Qty         string `json:"qty" url:"qty"`
	MarketUnit  string `json:"marketUnit,omitempty" url:"marketUnit,omitempty"`
	OrderLinkID string `json:"orderLinkId,omitempty" url:"orderLinkId,omitempty"`
}
