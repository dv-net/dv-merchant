//nolint:tagliatelle
package models

import (
	"net/url"
)

type CancelSource string

const (
	CancelSourceTimeoutCanceledOrder       CancelSource = "timeout-canceled-order"
	CancelSourceCrossMarginFlSys           CancelSource = "cross-margin-fl-sys"
	CancelSourceIsolatedMarginFlSys        CancelSource = "isolated-margin-fl-sys"
	CancelSourceCoinListingDelisting       CancelSource = "coin-listing-delisting"
	CancelSourceAPI                        CancelSource = "api"
	CancelSourceUserActivelyCancelsWeb     CancelSource = "user-actively-cancels-order-web"
	CancelSourceUserActivelyCancelsIOS     CancelSource = "user-actively-cancels-order-ios"
	CancelSourceUserActivelyCancelsAndroid CancelSource = "user-actively-cancels-order-ios"
	CancelSourceAdmin                      CancelSource = "admin"
	CancelSourceGridEnd                    CancelSource = "grid-end"
	CancelSourceSystemManuallyCancelsOrder CancelSource = "system-manually-cancels-order"
	CancelSourceCircuit                    CancelSource = "circuit"
	CancelSourceSelfMatchPrevent           CancelSource = "self_match_prevent"
	CancelSourceMarket                     CancelSource = "market"
	CancelSourceFok                        CancelSource = "fok"
	CancelSourceIOC                        CancelSource = "ioc"
	CancelSourceLimitMaker                 CancelSource = "limit_maker"
)

type OrderState string

func (o OrderState) String() string { return string(o) }

func (o OrderState) EncodeValues(key string, v *url.Values) error {
	v.Add(key, o.String())
	return nil
}

const (
	OrderStateCreated         OrderState = "created"
	OrderStateSubmitted       OrderState = "submitted"
	OrderStatePartialFilled   OrderState = "partial-filled"
	OrderStateFilled          OrderState = "filled"
	OrderStatePartialCanceled OrderState = "partial-canceled"
	OrderStateCanceling       OrderState = "canceling"
	OrderStateCanceled        OrderState = "canceled"
)

type OrderType string

func (o OrderType) String() string { return string(o) }

func (o OrderType) EncodeValues(key string, v *url.Values) error {
	v.Add(key, o.String())
	return nil
}

const (
	OrderTypeBuyMarket        OrderType = "buy-market"
	OrderTypeSellMarket       OrderType = "sell-market"
	OrderTypeBuyLimit         OrderType = "buy-limit"
	OrderTypeSellLimit        OrderType = "sell-limit"
	OrderTypeBuyIOC           OrderType = "buy-ioc"
	OrderTypeSellIOC          OrderType = "sell-ioc"
	OrderTypeBuyStopLimit     OrderType = "buy-stop-limit"
	OrderTypeSellStopLimit    OrderType = "sell-stop-limit"
	OrderTypeBuyLimitFok      OrderType = "buy-limit-fok"
	OrderTypeSellLimitFok     OrderType = "sell-limit-fok"
	OrderTypeBuyStopLimitFok  OrderType = "buy-stop-limit-fok"
	OrderTypeSellStopLimitFok OrderType = "sell-stop-limit-fok"
)

type OrderSource string

const (
	OrderSourceSys             OrderSource = "sys"
	OrderSourceWeb             OrderSource = "web"
	OrderSourceAPI             OrderSource = "api"
	OrderSourceSpotAPI         OrderSource = "spot-api"
	OrderSourceApp             OrderSource = "app"
	OrderSourceFlSys           OrderSource = "fl-sys"
	OrderSourceFlMgt           OrderSource = "fl-mgt"
	OrderSourceSpotStop        OrderSource = "spot-stop"
	OrderSourceMarginStop      OrderSource = "margin-stop"
	OrderSourceSuperMarginStop OrderSource = "super-margin-stop"
	OrderSourceGridTradingSys  OrderSource = "grid-trading-sys"
)

type Order struct {
	ID              int64        `json:"id,omitempty"`
	ClientOrderID   string       `json:"client-order-id,omitempty"`
	Symbol          string       `json:"symbol,omitempty"`
	AccountID       int64        `json:"account-id,omitempty"`
	Amount          string       `json:"amount,omitempty"`
	Price           string       `json:"price,omitempty"`
	CreatedAt       int64        `json:"created-at,omitempty"`
	FinishedAt      int64        `json:"finished-at,omitempty"`
	CanceledAt      int64        `json:"canceled-at,omitempty"`
	Type            OrderType    `json:"type,omitempty"`
	FieldAmount     string       `json:"field-amount,omitempty"`
	FieldCashAmount string       `json:"field-cash-amount,omitempty"`
	FieldFees       string       `json:"field-fees,omitempty"`
	Source          OrderSource  `json:"source,omitempty"`
	CanceledSource  CancelSource `json:"canceled-source,omitempty"`
	State           OrderState   `json:"state,omitempty"`
	StopPrice       string       `json:"stop-price,omitempty"`
	Operator        string       `json:"operator,omitempty"`
}
