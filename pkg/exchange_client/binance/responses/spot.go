package responses

import "github.com/shopspring/decimal"

type TestNewOrderResponse struct {
	StandardCommissionForOrder struct {
		Maker string `json:"maker,omitempty"`
		Taker string `json:"taker,omitempty"`
	} `json:"standardCommissionForOrder,omitempty"`
	TaxCommissionForOrder struct {
		Maker string `json:"maker,omitempty"`
		Taker string `json:"taker,omitempty"`
	} `json:"taxCommissionForOrder,omitempty"`
	Discount struct {
		EnabledForAccount bool   `json:"enabledForAccount,omitempty"`
		EnabledForSymbol  bool   `json:"enabledForSymbol,omitempty"`
		DiscountAsset     string `json:"discountAsset,omitempty"`
		Discount          string `json:"discount,omitempty"`
	} `json:"discount,omitempty"`
}

type NewOrderResponse struct {
	Symbol                  string          `json:"symbol"`
	OrderId                 int             `json:"orderId"`
	OrderListId             int             `json:"orderListId"`
	ClientOrderId           string          `json:"clientOrderId"`
	TransactTime            int64           `json:"transactTime"`
	Price                   string          `json:"price"`
	OrigQty                 string          `json:"origQty"`
	ExecutedQty             decimal.Decimal `json:"executedQty"`
	OrigQuoteOrderQty       string          `json:"origQuoteOrderQty"`
	CummulativeQuoteQty     string          `json:"cummulativeQuoteQty"`
	Status                  string          `json:"status"`
	TimeInForce             string          `json:"timeInForce"`
	Type                    string          `json:"type"`
	Side                    string          `json:"side"`
	WorkingTime             int64           `json:"workingTime"`
	SelfTradePreventionMode string          `json:"selfTradePreventionMode"`
	Fills                   []struct {
		Price           string `json:"price"`
		Qty             string `json:"qty"`
		Commission      string `json:"commission"`
		CommissionAsset string `json:"commissionAsset"`
		TradeId         int    `json:"tradeId"`
	} `json:"fills"`
}

type QueryOrderResponse struct {
	Symbol                  string `json:"symbol"`
	OrderId                 int    `json:"orderId"`
	OrderListId             int    `json:"orderListId"`
	ClientOrderId           string `json:"clientOrderId"`
	Price                   string `json:"price"`
	OrigQty                 string `json:"origQty"`
	ExecutedQty             string `json:"executedQty"`
	CummulativeQuoteQty     string `json:"cummulativeQuoteQty"`
	Status                  string `json:"status"`
	TimeInForce             string `json:"timeInForce"`
	Type                    string `json:"type"`
	Side                    string `json:"side"`
	StopPrice               string `json:"stopPrice"`
	IcebergQty              string `json:"icebergQty"`
	Time                    int64  `json:"time"`
	UpdateTime              int64  `json:"updateTime"`
	IsWorking               bool   `json:"isWorking"`
	WorkingTime             int64  `json:"workingTime"`
	OrigQuoteOrderQty       string `json:"origQuoteOrderQty"`
	SelfTradePreventionMode string `json:"selfTradePreventionMode"`
}

type CancelOrderResponse struct {
	Symbol                  string `json:"symbol"`
	OrigClientOrderId       string `json:"origClientOrderId"`
	OrderId                 int    `json:"orderId"`
	OrderListId             int    `json:"orderListId"`
	ClientOrderId           string `json:"clientOrderId"`
	TransactTime            int64  `json:"transactTime"`
	Price                   string `json:"price"`
	OrigQty                 string `json:"origQty"`
	ExecutedQty             string `json:"executedQty"`
	CummulativeQuoteQty     string `json:"cummulativeQuoteQty"`
	Status                  string `json:"status"`
	TimeInForce             string `json:"timeInForce"`
	Type                    string `json:"type"`
	Side                    string `json:"side"`
	SelfTradePreventionMode string `json:"selfTradePreventionMode"`
}

type AccountInformationResponse struct {
	CanTrade    bool `json:"canTrade"`
	CanWithdraw bool `json:"canWithdraw"`
	CanDeposit  bool `json:"canDeposit"`
}
