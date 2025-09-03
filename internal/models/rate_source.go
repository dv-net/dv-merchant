package models

type RateSource string

const (
	RateSourceOKX     RateSource = "okx"
	RateSourceHTX     RateSource = "htx"
	RateSourceBinance RateSource = "binance"
	RateSourceBitGet  RateSource = "bitget"
	RateSourceDVMin   RateSource = "dv-min"
	RateSourceDVMax   RateSource = "dv-max"
	RateSourceDVAvg   RateSource = "dv-avg"
	RateSourceKucoin  RateSource = "kucoin"
	RateSourceBybit   RateSource = "bybit"
	RateSourceGateio  RateSource = "gate"
)

func (rs RateSource) String() string {
	return string(rs)
}

func (rs RateSource) Valid() bool {
	switch rs {
	case RateSourceBinance:
		return true
	case RateSourceHTX:
		return true
	case RateSourceOKX:
		return true
	case RateSourceBitGet:
		return true
	case RateSourceDVMin:
		return true
	case RateSourceDVMax:
		return true
	case RateSourceDVAvg:
		return true
	case RateSourceKucoin:
		return true
	case RateSourceBybit:
		return true
	case RateSourceGateio:
		return true
	}

	return false
}
