package converters

import (
	"slices"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/public_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/store"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"
)

func ConvertTopUpDataToResponse(data *store.TopUpData) public_request.GetWalletDto {
	addresses := make([]public_request.WalletAddressDto, len(data.Addresses))
	for idx, address := range data.Addresses {
		wa := public_request.WalletAddressDto{
			Currency: public_request.CurrencyDto{
				ID:         address.CurrencyID,
				Blockchain: address.Blockchain,
			},
			Address: address.Address,
		}

		currencyIdx := slices.IndexFunc(data.AvailableCurrencies, func(cur *models.Currency) bool {
			return cur.ID == address.CurrencyID
		})

		if currencyIdx >= 0 {
			currency := data.AvailableCurrencies[currencyIdx]
			wa.Currency.Name = currency.Name
			wa.Currency.Code = currency.Code
			wa.Currency.CurrencyLabel = pgtypeutils.DecodeText(currency.CurrencyLabel)
			wa.Currency.TokenLabel = pgtypeutils.DecodeText(currency.TokenLabel)
			wa.Currency.ContractAddress = pgtypeutils.DecodeText(currency.ContractAddress)
			wa.Currency.IsNative = currency.IsNative
			wa.Currency.Order = currency.OrderIdx
		}

		addresses[idx] = wa
	}

	result := public_request.GetWalletDto{
		Store: public_request.StoreDto{
			ID:             data.Store.ID,
			Name:           data.Store.Name,
			CurrencyID:     data.Store.CurrencyID,
			SiteURL:        data.Store.Site,
			ReturnURL:      data.Store.ReturnUrl,
			SuccessURL:     data.Store.SuccessUrl,
			MinimalPayment: data.Store.MinimalPayment.String(),
			Status:         data.Store.Status,
		},
		WalletID:  data.WalletID,
		Addresses: addresses,
		Rates:     data.Rates,
	}

	return result
}
