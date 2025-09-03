package dictionary_response

import (
	"github.com/dv-net/dv-merchant/internal/models"
)

type GetDictionaryResponse struct {
	AvailableCurrencies  []*models.CurrencyShort `json:"available_currencies"`
	AvailableRateSources []string                `json:"available_rate_sources"`
} // @name GetDictionaryResponse
