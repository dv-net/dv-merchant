package converters

import (
	"slices"

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
	models.BlockchainArbitrum:          "https://blockchair.com/arbitrum-one/transaction/{tx_hash}",
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

func FromCurrencyModelsToExtendedResponse(currencies []*models.Currency) *currency_response.CurrenciesExtendedResponse {
	tokensMap := make(map[string]*currency_response.TokenGroup)
	blockchainsMap := make(map[string]*currency_response.BlockchainGroup)
	currencyItems := make([]*currency_response.CurrencyExtendedItem, 0, len(currencies))

	for _, m := range currencies {
		blockchain := ""
		if m.Blockchain != nil {
			blockchain = m.Blockchain.String()
		}

		// Build currency item
		item := &currency_response.CurrencyExtendedItem{
			ID:              m.ID,
			Code:            m.Code,
			Name:            m.Name,
			Blockchain:      blockchain,
			ContractAddress: m.ContractAddress.String,
			HasBalance:      m.HasBalance,
			MinConfirmation: int(m.MinConfirmation.Int16),
			Icon: currency_response.CurrencyIcon{
				Icon128: adminCDN + "img/currency/128/" + m.ID + ".png",
				Icon512: adminCDN + "img/currency/512/" + m.ID + ".png",
				IconSVG: adminCDN + "img/currency/svg/" + m.ID + ".svg",
			},
			TokenIcon: currency_response.CurrencyIcon{
				Icon128: adminCDN + "img/currency/128/" + m.Code + ".png",
				Icon512: adminCDN + "img/currency/512/" + m.Code + ".png",
				IconSVG: adminCDN + "img/currency/svg/" + m.Code + ".svg",
			},
		}

		if m.Blockchain != nil {
			if explLink, ok := blockchainExplorerDomainMap[*m.Blockchain]; ok {
				item.ExplorerLink = explLink
			}
			item.BlockchainIcon = currency_response.CurrencyIcon{
				Icon128: adminCDN + "img/blockchain/128/" + blockchain + ".png",
				Icon512: adminCDN + "img/blockchain/512/" + blockchain + ".png",
				IconSVG: adminCDN + "img/blockchain/svg/" + blockchain + ".svg",
			}
		}

		currencyItems = append(currencyItems, item)

		// Group by token (code)
		if _, exists := tokensMap[m.Code]; !exists {
			tokensMap[m.Code] = &currency_response.TokenGroup{
				Name: m.Code,
				Icon: currency_response.CurrencyIcon{
					Icon128: adminCDN + "img/currency/128/" + m.Code + ".png",
					Icon512: adminCDN + "img/currency/512/" + m.Code + ".png",
					IconSVG: adminCDN + "img/currency/svg/" + m.Code + ".svg",
				},
				Currencies:  make([]string, 0),
				Blockchains: make([]string, 0),
			}
		}
		tokenGroup := tokensMap[m.Code]
		tokenGroup.Currencies = append(tokenGroup.Currencies, m.ID)
		if blockchain != "" && !slices.Contains(tokenGroup.Blockchains, blockchain) {
			tokenGroup.Blockchains = append(tokenGroup.Blockchains, blockchain)
		}

		// Group by blockchain
		if blockchain != "" {
			if _, exists := blockchainsMap[blockchain]; !exists {
				blockchainsMap[blockchain] = &currency_response.BlockchainGroup{
					Name: blockchain,
					Icon: currency_response.CurrencyIcon{
						Icon128: adminCDN + "img/blockchain/128/" + blockchain + ".png",
						Icon512: adminCDN + "img/blockchain/512/" + blockchain + ".png",
						IconSVG: adminCDN + "img/blockchain/svg/" + blockchain + ".svg",
					},
					Currencies: make([]string, 0),
					Tokens:     make([]string, 0),
				}
			}
			blockchainGroup := blockchainsMap[blockchain]
			blockchainGroup.Currencies = append(blockchainGroup.Currencies, m.ID)
			if !slices.Contains(blockchainGroup.Tokens, m.Code) {
				blockchainGroup.Tokens = append(blockchainGroup.Tokens, m.Code)
			}
		}
	}

	// Convert maps to slices
	tokens := make([]*currency_response.TokenGroup, 0, len(tokensMap))
	for _, t := range tokensMap {
		tokens = append(tokens, t)
	}

	blockchains := make([]*currency_response.BlockchainGroup, 0, len(blockchainsMap))
	for _, b := range blockchainsMap {
		blockchains = append(blockchains, b)
	}

	return &currency_response.CurrenciesExtendedResponse{
		Tokens:      tokens,
		Blockchains: blockchains,
		Currencies:  currencyItems,
	}
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
