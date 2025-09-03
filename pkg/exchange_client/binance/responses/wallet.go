package responses

import binancemodels "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/models"

type GetDefaultDepositAddressResponse struct {
	Address string `json:"address"`
	Coin    string `json:"coin"`
	Tag     string `json:"tag,omitempty"`
	URL     string `json:"url"`
}

type WithdrawalResponse struct {
	ID string `json:"id"`
}

type GetAllCoinInformationResponse struct {
	Data []*binancemodels.CoinInfo
}

type GetFundingAssetsResponse struct {
	Data []binancemodels.AssetBalance
}

type GetSpotAssetsResponse struct {
	Data []binancemodels.AssetBalance
}

type GetDepositAddressesResponse struct {
	Data []struct {
		Coin      string `json:"coin"`
		Address   string `json:"address"`
		Tag       string `json:"tag"`
		IsDefault int    `json:"isDefault"`
	}
}

type GetUserBalancesResponse struct {
	Data []struct {
		Activate   bool   `json:"activate"`
		Balance    string `json:"balance"`
		WalletName string `json:"walletName"`
	}
}

type GetWithdrawalAddressesResponse struct {
	Data []binancemodels.WithdrawalWallet
}

type GetWithdrawalHistoryResponse struct {
	Data []binancemodels.WithdrawalInfo
}

type UniversalTransferResponse struct {
	TransactionID int64 `json:"tranId"`
}
