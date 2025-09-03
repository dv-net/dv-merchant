package requests

import binancemodels "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/models"

type TestNewOrderRequest struct {
	NewOrderRequest
	ComputeCommissionRates bool `json:"computeCommissionRates,omitempty" url:"computeCommissionRates,omitempty"`
}

type NewOrderRequest struct {
	Symbol                  string `json:"symbol" url:"symbol" validate:"required"`
	Side                    string `json:"side" url:"side" validate:"required"`
	Type                    string `json:"type" url:"type" validate:"required"`
	TimeInForce             string `json:"timeInForce,omitempty" url:"timeInForce,omitempty"`
	Quantity                string `json:"quantity,omitempty" url:"quantity,omitempty" validate:"required_without=QuoteOrderQty"`
	QuoteOrderQty           string `json:"quoteOrderQty,omitempty" url:"quoteOrderQty,omitempty" validate:"required_without=Quantity"`
	Price                   string `json:"price,omitempty" url:"price,omitempty"`
	NewClientOrderId        string `json:"newClientOrderId,omitempty" url:"newClientOrderId,omitempty"`
	StrategyId              int    `json:"strategyId,omitempty" url:"strategyId,omitempty"`
	StrategyType            int    `json:"strategyType,omitempty" url:"strategyType,omitempty"`
	StopPrice               string `json:"stopPrice,omitempty" url:"stopPrice,omitempty"`
	TrailingDelta           int    `json:"trailingDelta,omitempty" url:"trailingDelta,omitempty"`
	IcebergQty              string `json:"icebergQty,omitempty" url:"icebergQty,omitempty"`
	NewOrderRespType        string `json:"newOrderRespType,omitempty" url:"newOrderRespType,omitempty"`
	SelfTradePreventionMode string `json:"selfTradePreventionMode,omitempty" url:"selfTradePreventionMode,omitempty"`
	RecvWindow              int64  `json:"-" url:"-"`
	Timestamp               int    `json:"-" url:"-"`
}

type StrictNewOrderRequest struct {
	Symbol                  string `json:"symbol" url:"symbol" validate:"required"`
	Side                    string `json:"side" url:"side" validate:"required"`
	Type                    string `json:"type" url:"type" validate:"required"`
	TimeInForce             string `json:"timeInForce,omitempty" url:"timeInForce,omitempty" validate:"required_if=Type LIMIT required_if=Type STOP_LOSS_LIMIT required_if=Type TAKE_PROFIT_LIMIT"`
	Quantity                string `json:"quantity,omitempty" url:"quantity,omitempty" validate:"required_without=QuoteOrderQty required_if=Type LIMIT required_if=Type STOP_LOSS required_if=Type STOP_LOSS_LIMIT required_if=Type TAKE_PROFIT required_if=Type TAKE_PROFIT_LIMIT required_if=Type LIMIT_MAKER"`
	QuoteOrderQty           string `json:"quoteOrderQty,omitempty" url:"quoteOrderQty,omitempty" validate:"required_without=Quantity required_if=Type MARKET"`
	Price                   string `json:"price,omitempty" url:"price,omitempty" validate:"required_if=Type LIMIT required_if=Type STOP_LOSS_LIMIT required_if=Type TAKE_PROFIT_LIMIT required_if=Type LIMIT_MAKER"`
	NewClientOrderId        string `json:"newClientOrderId,omitempty" url:"newClientOrderId,omitempty"`
	StrategyId              int    `json:"strategyId,omitempty" url:"strategyId,omitempty"`
	StrategyType            int    `json:"strategyType,omitempty" url:"strategyType,omitempty"`
	StopPrice               string `json:"stopPrice,omitempty" url:"stopPrice,omitempty" validate:"required_if=Type STOP_LOSS required_if=Type STOP_LOSS_LIMIT required_if=Type TAKE_PROFIT required_if=Type TAKE_PROFIT_LIMIT"`
	TrailingDelta           int    `json:"trailingDelta,omitempty" url:"trailingDelta,omitempty" validate:"required_if=Type STOP_LOSS required_if=Type STOP_LOSS_LIMIT required_if=Type TAKE_PROFIT required_if=Type TAKE_PROFIT_LIMIT"`
	IcebergQty              string `json:"icebergQty,omitempty" url:"icebergQty,omitempty"`
	NewOrderRespType        string `json:"newOrderRespType,omitempty" url:"newOrderRespType,omitempty"`
	SelfTradePreventionMode string `json:"selfTradePreventionMode,omitempty" url:"selfTradePreventionMode,omitempty"`
	RecvWindow              int64  `json:"-" url:"-"`
	Timestamp               int64  `json:"-" url:"-"`
}

type QueryOrderRequest struct {
	Symbol            string `json:"symbol" url:"symbol" validate:"required"`
	OrderID           int64  `json:"orderId" url:"orderId" validate:"required_without=OrigClientOrderId"`
	OrigClientOrderId string `json:"origClientOrderId" url:"origClientOrderId" validate:"required_without=OrderID"`
	RecvWindow        int64  `json:"-" url:"-"`
	Timestamp         int64  `json:"-" url:"-"`
}

type CancelOrderRequest struct {
	Symbol             string                           `json:"symbol" url:"symbol" validate:"required"`
	OrderID            int64                            `json:"orderId,omitempty" url:"orderId,omitempty"`
	OrigClientOrderId  string                           `json:"origClientOrderId,omitempty" url:"origClientOrderId,omitempty"`
	NewClientOrderId   string                           `json:"newClientOrderId,omitempty" url:"newClientOrderId,omitempty"`
	CancelRestrictions binancemodels.CancelRestrictions `json:"cancelRestrictions,omitempty" url:"cancelRestrictions,omitempty"`
	RecvWindow         int64                            `json:"-" url:"-"`
	Timestamp          int64                            `json:"-" url:"-"`
}

type AccountInformationRequest struct {
	RecvWindow int64 `json:"-" url:"-"`
	Timestamp  int64 `json:"-" url:"-"`
}
