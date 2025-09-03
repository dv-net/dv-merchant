//nolint:tagliatelle
package responses

import "github.com/dv-net/dv-merchant/pkg/exchange_client/bybit/models"

type GetActiveOrdersResponse struct {
	Category string         `json:"category"`
	List     []models.Order `json:"list"`
}

type GetOrderHistoryResponse struct {
	Category string         `json:"category"`
	List     []models.Order `json:"list"`
}

type PlaceOrderResponse struct {
	OrderID     string `json:"orderId"`
	OrderLinkID string `json:"orderLinkId,omitempty"`
}
