package models

type ExchangeSlug string // @name ExchangeSlug

const (
	ExchangeSlugHtx     ExchangeSlug = "htx"
	ExchangeSlugOkx     ExchangeSlug = "okx"
	ExchangeSlugBinance ExchangeSlug = "binance"
	ExchangeSlugBitget  ExchangeSlug = "bitget"
	ExchangeSlugKucoin  ExchangeSlug = "kucoin"
	ExchangeSlugBybit   ExchangeSlug = "bybit"
	ExchangeSlugGateio  ExchangeSlug = "gate"
)

func (o ExchangeSlug) Valid() bool {
	switch o {
	case ExchangeSlugHtx:
		return true
	case ExchangeSlugOkx:
		return true
	case ExchangeSlugBinance:
		return true
	case ExchangeSlugBitget:
		return true
	case ExchangeSlugKucoin:
		return true
	case ExchangeSlugBybit:
		return true
	case ExchangeSlugGateio:
		return true
	default:
		return false
	}
}

func (o ExchangeSlug) String() string {
	return string(o)
}
