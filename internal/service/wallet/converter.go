package wallet

import (
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/dv-net/dv-processing/pkg/walletsdk/bch"
	"github.com/gcash/bchd/chaincfg"
)

type IWalletAddressConverter interface {
	ConvertLegacyAddressToNewFormat(address string, blockchain models.Blockchain) (*ConvertAddressDTO, error)
}

func (s *Service) ConvertLegacyAddressToNewFormat(address string, blockchain models.Blockchain) (*ConvertAddressDTO, error) {
	if blockchain != models.BlockchainBitcoinCash {
		return nil, fmt.Errorf("conversion not required for blockchain: %s", blockchain)
	}

	params := chaincfg.MainNetParams
	var response ConvertAddressDTO

	if isLegacy, err := bch.IsLegacyAddress(address, &params); err != nil {
		return nil, err
	} else if isLegacy {
		cashAddr, err := bch.DecodeAddressToCashAddr(address, &params)
		if err != nil {
			return nil, err
		}
		response.LegacyAddress = &address
		response.Address = &cashAddr
		return &response, nil
	}

	if isCashAddr, err := bch.IsCashAddrAddress(address, &params); err != nil {
		return nil, err
	} else if isCashAddr {
		legacyAddr, err := bch.DecodeAddressToLegacyAddr(address, &params)
		if err != nil {
			return nil, err
		}
		response.LegacyAddress = &legacyAddr
		response.Address = &address
		return &response, nil
	}

	return nil, fmt.Errorf("invalid Bitcoin Cash address format: %s", address)
}
