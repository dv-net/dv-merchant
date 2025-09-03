package responses

import "github.com/shopspring/decimal"

type Result[T InitAuthResponse | OwnerDataResponse | InitOwnerTgResponse | CheckMyIPResponse | []RatesResponse] struct {
	Data    T      `json:"data"`
	Message string `json:"message"`
}

type InitAuthResponse struct {
	Link  string `json:"link"`
	Token string `json:"token"`
}

type OwnerDataResponse struct {
	Telegram *string         `json:"telegram"`
	Balance  decimal.Decimal `json:"balance"`
}

type InitOwnerTgResponse struct {
	Link string `json:"link"`
}
