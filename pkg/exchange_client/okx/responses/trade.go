//nolint:tagliatelle
package responses

import (
	okxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/models"
)

type (
	PlaceOrder struct {
		Basic
		PlaceOrders []*okxmodels.PlaceOrder `json:"data"`
	}
	CancelOrder struct {
		Basic
		CancelOrders []*okxmodels.CancelOrder `json:"data"`
	}
	AmendOrder struct {
		Basic
		AmendOrders []*okxmodels.AmendOrder `json:"data"`
	}
	ClosePosition struct {
		Basic
		ClosePositions []*okxmodels.ClosePosition `json:"data"`
	}
	OrderList struct {
		Basic
		Orders []*okxmodels.Order `json:"data"`
	}
	TransactionDetail struct {
		Basic
		TransactionDetails []*okxmodels.TransactionDetail `json:"data"`
	}
	PlaceAlgoOrder struct {
		Basic
		PlaceAlgoOrders []*okxmodels.PlaceAlgoOrder `json:"data"`
	}
	CancelAlgoOrder struct {
		Basic
		CancelAlgoOrders []*okxmodels.CancelAlgoOrder `json:"data"`
	}
	AlgoOrderList struct {
		Basic
		AlgoOrders []*okxmodels.AlgoOrder `json:"data"`
	}
)
