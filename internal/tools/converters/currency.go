package converters

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/currency_response"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/store_response"
	"github.com/dv-net/dv-merchant/internal/models"
)

const adminCDN = "https://cdn.dv.net/"

var blockchainExplorerDomainMap = map[models.Blockchain]string{
	models.BlockchainEthereum:          "https://etherscan.io/tx/{tx_hash}",
	models.BlockchainBinanceSmartChain: "https://bscscan.com/tx/{tx_hash}",
	models.BlockchainBitcoin:           "https://blockchair.com/bitcoin/transaction/{tx_hash}",
	models.BlockchainTron:              "https://tronscan.org/#/transaction/{tx_hash}",
	models.BlockchainBitcoinCash:       "https://blockchair.com/bitcoin-cash/transaction/{tx_hash}",
	models.BlockchainLitecoin:          "https://blockchair.com/litecoin/transaction/{tx_hash}",
	models.BlockchainPolygon:           "https://polygonscan.com/tx/{tx_hash}",
	models.BlockchainDogecoin:          "https://blockchair.com/dogecoin/transaction/{tx_hash}",
}

func FromCurrencyModelToResponse(model *models.Currency) *currency_response.GetCurrencyResponse {
	c := &currency_response.GetCurrencyResponse{
		ID:              model.ID,
		Code:            model.Code,
		Name:            model.Name,
		IsFiat:          model.IsFiat,
		Precision:       int(model.Precision),
		ContractAddress: model.ContractAddress.String,
		HasBalance:      model.HasBalance,
		Status:          model.Status,
		MinConfirmation: int(model.MinConfirmation.Int16),
	}
	if model.Blockchain != nil {
		c.Blockchain = model.Blockchain.String()
	}
	if model.WithdrawalMinBalance != nil {
		c.WithdrawalMinBalance = *model.WithdrawalMinBalance
	}
	currencyIcons := &currency_response.CurrencyIcon{
		Icon128: adminCDN + "img/currency/128/" + model.ID + ".png",
		Icon512: adminCDN + "img/currency/512/" + model.ID + ".png",
		IconSVG: adminCDN + "img/currency/svg/" + model.ID + ".svg",
	}
	c.Icon = *currencyIcons

	if model.Blockchain != nil {
		explLink, ok := blockchainExplorerDomainMap[*model.Blockchain]
		if ok {
			c.ExplorerLink = explLink
		}
		blockchainIcons := &currency_response.CurrencyIcon{
			Icon128: adminCDN + "img/blockchain/128/" + model.Blockchain.String() + ".png",
			Icon512: adminCDN + "img/blockchain/512/" + model.Blockchain.String() + ".png",
			IconSVG: adminCDN + "img/blockchain/svg/" + model.Blockchain.String() + ".svg",
		}
		c.BlockchainIcon = *blockchainIcons
	}

	return c
}

func FromCurrencyModelToResponses(models ...*models.Currency) []*currency_response.GetCurrencyResponse {
	res := make([]*currency_response.GetCurrencyResponse, 0, len(models))
	for _, model := range models {
		res = append(res, FromCurrencyModelToResponse(model))
	}
	return res
}

func FromStoreCurrencyModelToResponse(m *models.StoreCurrency) *store_response.StoreCurrencyResponse {
	return &store_response.StoreCurrencyResponse{m.CurrencyID}
}

func FromStoreCurrencyModelToResponses(models ...*models.StoreCurrency) []*store_response.StoreCurrencyResponse {
	res := make([]*store_response.StoreCurrencyResponse, 0, len(models))
	for _, model := range models {
		res = append(res, FromStoreCurrencyModelToResponse(model))
	}
	return res
}
