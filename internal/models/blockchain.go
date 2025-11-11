package models

import (
	"fmt"

	commonv1 "github.com/dv-net/dv-processing/api/processing/common/v1"
	commonv2 "github.com/dv-net/dv-proto/gen/go/eproxy/common/v2"
)

var BlockchainSortOrder = map[Blockchain]int{
	BlockchainTron:              0,
	BlockchainEthereum:          1,
	BlockchainBinanceSmartChain: 2,
	BlockchainPolygon:           3,
	BlockchainBitcoin:           4,
	BlockchainLitecoin:          5,
	BlockchainBitcoinCash:       6,
	BlockchainDogecoin:          7,
	BlockchainArbitrum:          8,
}

func (b Blockchain) ToPb() (commonv1.Blockchain, error) {
	switch b {
	case BlockchainBitcoin:
		return commonv1.Blockchain_BLOCKCHAIN_BITCOIN, nil
	case BlockchainTron:
		return commonv1.Blockchain_BLOCKCHAIN_TRON, nil
	case BlockchainEthereum:
		return commonv1.Blockchain_BLOCKCHAIN_ETHEREUM, nil
	case BlockchainLitecoin:
		return commonv1.Blockchain_BLOCKCHAIN_LITECOIN, nil
	case BlockchainBitcoinCash:
		return commonv1.Blockchain_BLOCKCHAIN_BITCOINCASH, nil
	case BlockchainBinanceSmartChain:
		return commonv1.Blockchain_BLOCKCHAIN_BINANCE_SMART_CHAIN, nil
	case BlockchainPolygon:
		return commonv1.Blockchain_BLOCKCHAIN_POLYGON, nil
	case BlockchainDogecoin:
		return commonv1.Blockchain_BLOCKCHAIN_DOGECOIN, nil
	case BlockchainArbitrum:
		return commonv1.Blockchain_BLOCKCHAIN_ARBITRUM, nil
	}
	return commonv1.Blockchain_BLOCKCHAIN_UNSPECIFIED, fmt.Errorf("invalid blockchain: %s", b)
}

func (b Blockchain) ToEPb() (commonv2.Blockchain, error) {
	switch b {
	case BlockchainBitcoin:
		return commonv2.Blockchain_BLOCKCHAIN_BITCOIN, nil
	case BlockchainTron:
		return commonv2.Blockchain_BLOCKCHAIN_TRON, nil
	case BlockchainEthereum:
		return commonv2.Blockchain_BLOCKCHAIN_ETHEREUM, nil
	case BlockchainBitcoinCash:
		return commonv2.Blockchain_BLOCKCHAIN_BITCOINCASH, nil
	case BlockchainBinanceSmartChain:
		return commonv2.Blockchain_BLOCKCHAIN_BINANCE_SMART_CHAIN, nil
	case BlockchainLitecoin:
		return commonv2.Blockchain_BLOCKCHAIN_LITECOIN, nil
	case BlockchainPolygon:
		return commonv2.Blockchain_BLOCKCHAIN_POLYGON, nil
	case BlockchainDogecoin:
		return commonv2.Blockchain_BLOCKCHAIN_DOGECOIN, nil
	case BlockchainArbitrum:
		return commonv2.Blockchain_BLOCKCHAIN_ARBITRUM, nil
	default:
		return commonv2.Blockchain_BLOCKCHAIN_UNSPECIFIED, fmt.Errorf("invalid blockchain: %s", b)
	}
}

func ConvertToModel(blockchain commonv1.Blockchain) Blockchain {
	switch blockchain {
	case commonv1.Blockchain_BLOCKCHAIN_BITCOIN:
		return BlockchainBitcoin
	case commonv1.Blockchain_BLOCKCHAIN_TRON:
		return BlockchainTron
	case commonv1.Blockchain_BLOCKCHAIN_ETHEREUM:
		return BlockchainEthereum
	case commonv1.Blockchain_BLOCKCHAIN_LITECOIN:
		return BlockchainLitecoin
	case commonv1.Blockchain_BLOCKCHAIN_BITCOINCASH:
		return BlockchainBitcoinCash
	case commonv1.Blockchain_BLOCKCHAIN_BINANCE_SMART_CHAIN:
		return BlockchainBinanceSmartChain
	case commonv1.Blockchain_BLOCKCHAIN_POLYGON:
		return BlockchainPolygon
	case commonv1.Blockchain_BLOCKCHAIN_DOGECOIN:
		return BlockchainDogecoin
	case commonv1.Blockchain_BLOCKCHAIN_ARBITRUM:
		return BlockchainArbitrum
	}
	return ""
}

func (b Blockchain) RecalculateNativeBalance() bool {
	switch b {
	case BlockchainTron, BlockchainEthereum, BlockchainBinanceSmartChain, BlockchainPolygon, BlockchainArbitrum:
		return true
	default:
		return false
	}
}

func (b Blockchain) KindWithdrawalRequired() bool {
	switch b {
	case BlockchainTron:
		return true
	default:
		return false
	}
}

func (b Blockchain) ShowProcessingWallets() bool {
	switch b {
	case BlockchainEthereum, BlockchainTron, BlockchainBinanceSmartChain, BlockchainPolygon, BlockchainArbitrum:
		return true
	default:
		return false
	}
}

func (b Blockchain) IsBitcoinLike() bool {
	switch b {
	case BlockchainBitcoin, BlockchainLitecoin, BlockchainBitcoinCash, BlockchainDogecoin:
		return true
	default:
		return false
	}
}

func (b Blockchain) IsEVMLike() bool {
	switch b {
	case BlockchainEthereum, BlockchainBinanceSmartChain, BlockchainPolygon, BlockchainArbitrum:
		return true
	default:
		return false
	}
}

func (b Blockchain) ConfirmationBlockCount() uint64 {
	switch b {
	case BlockchainBitcoin:
		return 1
	case BlockchainLitecoin:
		return 3
	case BlockchainBinanceSmartChain:
		return 15
	case BlockchainEthereum:
		return 32
	case BlockchainTron:
		return 19
	case BlockchainDogecoin:
		return 12
	case BlockchainArbitrum:
		return 6
	case BlockchainPolygon:
		return 300
	default:
		return 0
	}
}
