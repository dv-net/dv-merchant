package chains

import (
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"
)

var HtxSupportedChains = map[string]struct{}{
	"trc20usdt": {},
	"usdterc20": {},
	"trx1":      {},
	"btc":       {},
	"eth":       {},
	"usdc":      {},
}

var OkxSupportedChains = map[string]struct{}{
	"BTC-Bitcoin": {},
	"ETH-ERC20":   {},
	"TRX-TRC20":   {},
	"USDT-TRC20":  {},
	"USDT-ERC20":  {},
	"USDC-ERC20":  {},
}

var HtxChainToCurrency = map[string]string{
	"btc":       "BTC.Bitcoin",
	"eth":       "ETH.Ethereum",
	"trx1":      "TRX.Tron",
	"trc20usdt": "USDT.Tron",
	"usdterc20": "USDT.Ethereum",
	"usdc":      "USDC.Ethereum",
}

var OkxChainToCurrency = map[string]string{
	"BTC-Bitcoin": "BTC.Bitcoin",
	"ETH-ERC20":   "ETH.Ethereum",
	"TRX-TRC20":   "TRX.Tron",
	"USDT-TRC20":  "USDT.Tron",
	"USDT-ERC20":  "USDT.Ethereum",
	"USDC-ERC20":  "USDC.Ethereum",
}

var OkxCurrencies = map[string]string{
	"BTC.Bitcoin":   "BTC",
	"ETH.Ethereum":  "ETH",
	"TRX.Tron":      "TRX",
	"USDT.Tron":     "USDT",
	"USDT.Ethereum": "USDT",
	"USDC.Ethereum": "USDC",
}

func GetCurrencyByID(exchange models.ExchangeSlug, id string) (string, error) {
	switch exchange {
	case models.ExchangeSlugHtx:
	case models.ExchangeSlugOkx:
		for k, v := range OkxCurrencies {
			if k == id {
				return v, nil
			}
		}
	default:
		return "", fmt.Errorf("currency not found for id: %s", id)
	}
	return "", fmt.Errorf("currency not found for id: %s", id)
}

func GetIDByCurrency(exchange models.ExchangeSlug, currency string) (string, error) {
	switch exchange {
	case models.ExchangeSlugHtx:
	case models.ExchangeSlugOkx:
		for k, v := range OkxCurrencies {
			if v == currency {
				return k, nil
			}
		}
	default:
		return "", fmt.Errorf("currency not found for currency: %s", currency)
	}
	return "", fmt.Errorf("currency not found for currency: %s", currency)
}

func GetCurrencyBySlugAndChain(exchange, chain string) (string, error) {
	switch exchange {
	case models.ExchangeSlugHtx.String():
		if c, ok := HtxChainToCurrency[chain]; ok {
			return c, nil
		}
	case models.ExchangeSlugOkx.String():
		if c, ok := OkxChainToCurrency[chain]; ok {
			return c, nil
		}
	default:
		return "", fmt.Errorf("currency not found for chain: %s", chain)
	}

	return "", fmt.Errorf("currency not found for chain: %s", chain)
}

func IsChainSupported(chain string) bool {
	_, htxExists := HtxSupportedChains[chain]
	_, okxExists := OkxSupportedChains[chain]
	return htxExists || okxExists
}
