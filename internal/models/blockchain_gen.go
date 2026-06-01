// Code generated. DO NOT EDIT.

package models

import "errors"

type Blockchain string // @name Blockchain

const (
	BlockchainArbitrum          Blockchain = "arbitrum"
	BlockchainBitcoin           Blockchain = "bitcoin"
	BlockchainBitcoinCash       Blockchain = "bitcoincash"
	BlockchainBinanceSmartChain Blockchain = "bsc"
	BlockchainDogecoin          Blockchain = "dogecoin"
	BlockchainEthereum          Blockchain = "ethereum"
	BlockchainLitecoin          Blockchain = "litecoin"
	BlockchainPolygon           Blockchain = "polygon"
	BlockchainTron              Blockchain = "tron"
)

func (o Blockchain) String() string { return string(o) }

func (o Blockchain) Valid() error {
	switch o {
	case BlockchainArbitrum:
		return nil
	case BlockchainBitcoin:
		return nil
	case BlockchainBitcoinCash:
		return nil
	case BlockchainBinanceSmartChain:
		return nil
	case BlockchainDogecoin:
		return nil
	case BlockchainEthereum:
		return nil
	case BlockchainLitecoin:
		return nil
	case BlockchainPolygon:
		return nil
	case BlockchainTron:
		return nil
	}
	return errors.New("invalid blockchain: " + string(o))
}

func (o Blockchain) Tokens() ([]string, error) {
	switch o {
	case BlockchainArbitrum:
		return []string{"USDT.Arbitrum", "USDC.Arbitrum", "DAI.Arbitrum", "ARB.Arbitrum", "CAKE.Arbitrum", "PYUSD.Arbitrum"}, nil
	case BlockchainBitcoin:
		return []string{}, nil
	case BlockchainBitcoinCash:
		return []string{}, nil
	case BlockchainBinanceSmartChain:
		return []string{"USDT.BNBSmartChain", "USDC.BNBSmartChain", "DAI.BNBSmartChain", "USDD.BNBSmartChain", "USDE.BNBSmartChain", "CAKE.BNBSmartChain", "SHIB.BNBSmartChain", "WLFI.BNBSmartChain", "USD1.BNBSmartChain"}, nil
	case BlockchainDogecoin:
		return []string{}, nil
	case BlockchainEthereum:
		return []string{"USDT.Ethereum", "USDC.Ethereum", "DAI.Ethereum", "USDD.Ethereum", "USDE.Ethereum", "ARB.Ethereum", "CAKE.Ethereum", "SHIB.Ethereum", "PEPE.Ethereum", "ENA.Ethereum", "WLFI.Ethereum", "USD1.Ethereum", "WLD.Ethereum", "PYUSD.Ethereum", "XAUT.Ethereum", "SAND.Ethereum"}, nil
	case BlockchainLitecoin:
		return []string{}, nil
	case BlockchainPolygon:
		return []string{"USDT.Polygon", "USDC.Polygon", "DAI.Polygon", "SAND.Polygon"}, nil
	case BlockchainTron:
		return []string{"USDT.Tron", "USDD.Tron", "USD1.Tron"}, nil

	}
	return nil, errors.New("invalid blockchain: " + string(o))
}

func (o Blockchain) NativeCurrency() (string, error) {
	switch o {
	case BlockchainArbitrum:
		return "ETH.Arbitrum", nil
	case BlockchainBitcoin:
		return "BTC.Bitcoin", nil
	case BlockchainBitcoinCash:
		return "BCH.Bitcoincash", nil
	case BlockchainBinanceSmartChain:
		return "BNB.BNBSmartChain", nil
	case BlockchainDogecoin:
		return "DOGE.Dogecoin", nil
	case BlockchainEthereum:
		return "ETH.Ethereum", nil
	case BlockchainLitecoin:
		return "LTC.Litecoin", nil
	case BlockchainPolygon:
		return "POL.Polygon", nil
	case BlockchainTron:
		return "TRX.Tron", nil

	}
	return "", errors.New("invalid blockchain: " + string(o))
}

func AllBlockchain() []Blockchain {
	return []Blockchain{
		BlockchainArbitrum,
		BlockchainBitcoin,
		BlockchainBitcoinCash,
		BlockchainBinanceSmartChain,
		BlockchainDogecoin,
		BlockchainEthereum,
		BlockchainLitecoin,
		BlockchainPolygon,
		BlockchainTron,
	}
}
