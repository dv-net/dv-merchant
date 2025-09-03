package utils

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	binancemodels "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/models"

	"github.com/shopspring/decimal"
)

func HashLimiterKey(entries ...string) string {
	h := sha256.New()
	h.Write([]byte(strings.Join(entries, "")))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func ConvertPrecision(withdrawIntegerMultiple string) decimal.Decimal {
	precision := decimal.Decimal{}
	if withdrawIntegerMultiple != "0" {
		p, found := strings.CutPrefix(withdrawIntegerMultiple, "0.")
		if found {
			precision = decimal.NewFromInt(int64(len(p)))
		}
	}
	return precision
}

func ExtractMarketFilters(filters []interface{}) (*binancemodels.MarketFilters, error) {
	marketFilters := &binancemodels.MarketFilters{}
	for _, filter := range filters {
		if f, ok := filter.(map[string]interface{}); ok {
			switch f["filterType"] {
			case "NOTIONAL":
				if err := i2s(filter, &marketFilters.NotionalFilter); err != nil {
					return nil, err
				}
			case "LOT_SIZE":
				if err := i2s(filter, &marketFilters.LotSizeFilter); err != nil {
					return nil, err
				}
			}
		}
	}
	return marketFilters, nil
}

func i2s(input interface{}, dest interface{}) error {
	jsonString, err := json.Marshal(input)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(jsonString, dest); err != nil {
		return err
	}
	return nil
}
