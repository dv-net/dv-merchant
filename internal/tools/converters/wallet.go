package converters

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/wallet_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/wallet"
)

func WalletBalanceModelToResponse(model *models.WalletWithUSDBalance) wallet_response.GetWalletBalanceResponse {
	return wallet_response.GetWalletBalanceResponse{
		ID:         model.WalletAddressID,
		CurrencyID: model.CurrencyID,
		Address:    model.Address,
		Blockchain: model.Blockchain,
		Amount:     model.Amount,
		AmountUSD:  model.AmountUSD,
		Dirty:      model.Dirty,
	}
}

func WalletBalanceModelsToResponses(models ...*models.WalletWithUSDBalance) []wallet_response.GetWalletBalanceResponse {
	balances := make([]wallet_response.GetWalletBalanceResponse, 0, len(models))
	for _, model := range models {
		walletBalance := WalletBalanceModelToResponse(model)
		balances = append(balances, walletBalance)
	}
	return balances
}

func FromWalletAddressModelToResponse(o *models.WalletAddress) *wallet_response.WalletAddressResponse {
	return &wallet_response.WalletAddressResponse{
		ID:         o.ID,
		WalletID:   o.AccountID.UUID,
		UserID:     o.UserID,
		CurrencyID: o.CurrencyID,
		Blockchain: o.Blockchain.String(),
		Address:    o.Address,
		CreatedAt:  o.CreatedAt.Time,
	}
}

func FromWalletAddressModelToResponses(o ...*models.WalletAddress) []*wallet_response.WalletAddressResponse {
	res := make([]*wallet_response.WalletAddressResponse, len(o))
	for i, v := range o {
		res[i] = FromWalletAddressModelToResponse(v)
	}
	return res
}

func FromWalletWithAddressModelToResponse(o *wallet.WithAddressDto) *wallet_response.CreateWalletExternalResponse {
	return &wallet_response.CreateWalletExternalResponse{
		ID:              o.ID,
		StoreID:         o.StoreID,
		StoreExternalID: o.StoreExternalID,
		Address:         FromWalletAddressModelToResponses(o.Address...),
		PayURL:          o.PayURL.String(),
		Rates:           o.Rates,
		AmountUSD:       o.AmountUSD,
		CreatedAt:       o.CreatedAt.Time,
		UpdatedAt:       o.UpdatedAt.Time,
	}
}
