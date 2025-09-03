package converters

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/dictionary_response"
	"github.com/dv-net/dv-merchant/internal/service/dictionary"
)

func FromDictionaryModelToResponse(d *dictionary.Dictionary) *dictionary_response.GetDictionaryResponse {
	return &dictionary_response.GetDictionaryResponse{
		AvailableRateSources: d.AvailableSources,
		AvailableCurrencies:  d.AvailableCurrencies,
	}
}
