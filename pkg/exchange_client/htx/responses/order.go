//nolint:tagliatelle
package responses

import (
	htxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/models"
)

type (
	GetOrder struct {
		Basic
		Order *htxmodels.Order `json:"data,omitempty"`
	}
	GetOrdersHistory struct {
		Basic
		Orders []*htxmodels.Order `json:"data,omitempty"`
	}
	PlaceOrder struct {
		Basic
		OrderID string `json:"data,omitempty"`
	}
)
