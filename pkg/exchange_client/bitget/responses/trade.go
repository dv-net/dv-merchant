package responses

import (
	bitgetmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/models"
)

type (
	OrderInformationResponse struct {
		CommonResponse
		Data []*bitgetmodels.OrderInformation `json:"data,omitempty"`
	}
	PlaceOrderResponse struct {
		CommonResponse
		Data *bitgetmodels.PlacedOrder `json:"data,omitempty"`
	}
)
