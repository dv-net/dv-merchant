package models

type CancelRestrictions string

func (o CancelRestrictions) String() string { return string(o) }

const (
	CancelRestrictionsOnlyNew             CancelRestrictions = "ONLY_NEW"
	CancelRestrictionsOnlyPartiallyFilled CancelRestrictions = "ONLY_PARTIALLY_FILLED"
)

type OrderSide string

func (o OrderSide) String() string { return string(o) }

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

type OrderType string

func (o OrderType) String() string { return string(o) }

const (
	OrderTypeMarket OrderType = "MARKET"
)

type OrderStatus string

func (o OrderStatus) String() string { return string(o) }

const (
	OrderStatusNew             OrderStatus = "NEW"
	OrderStatusPendingNew      OrderStatus = "PENDING_NEW"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCanceled        OrderStatus = "CANCELED"
	OrderStatusPendingCancel   OrderStatus = "PENDING_CANCEL"
	OrderStatusRejected        OrderStatus = "REJECTED"
	OrderStatusExpired         OrderStatus = "EXPIRED"
	OrderStatusExpiredInMatch  OrderStatus = "EXPIRED_IN_MATCH"
)
