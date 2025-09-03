package models

type OrderType string

func (o OrderType) String() string { return string(o) }

const (
	OrderTypeLimit  OrderType = "limit"
	OrderTypeMarket OrderType = "market"
)

type OrderStatus string

func (o OrderStatus) String() string { return string(o) }

const (
	OrderStatusLive            OrderStatus = "live"
	OrderStatusPartiallyFilled OrderStatus = "partially_filled"
	OrderStatusFilled          OrderStatus = "filled"
	OrderStatusCanceled        OrderStatus = "canceled"
)

type OrderSide string

func (o OrderSide) String() string { return string(o) }

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)
