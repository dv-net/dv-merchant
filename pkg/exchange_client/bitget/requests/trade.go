//nolint:tagliatelle
package requests

import "github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/models"

type (
	OrderInformationRequest struct {
		OrderID       string `json:"orderId,omitempty" url:"orderId,omitempty"`
		ClientOID     string `json:"clientOid,omitempty" url:"clientOid,omitempty"`
		RequestTime   string `json:"requestTime,omitempty" url:"requestTime,omitempty"`
		ReceiveWindow string `json:"receiveWindow,omitempty" url:"receiveWindow,omitempty"`
	}
	PlaceOrderRequest struct {
		Symbol    string           `json:"symbol"`
		Side      models.OrderSide `json:"side"`
		OrderType models.OrderType `json:"orderType"`
		Size      string           `json:"size,omitempty"`
		ClientOID string           `json:"clientOid,omitempty"`
	}
)
