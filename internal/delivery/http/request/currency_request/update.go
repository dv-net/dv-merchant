package currency_request

import "github.com/shopspring/decimal"

type UpdateCurrencyRateRequest struct {
	RateSource string          `json:"rate_source" binding:"required, oneof=okx htx binance bitget bybit gate dv-min dv-max dv-avg" enums:"okx htx binance bitget bybit gate dv-min dv-max dv-avg"`
	RateScale  decimal.Decimal `json:"rate_scale" binding:"required"`
}
