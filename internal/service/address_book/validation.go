package address_book

import (
	"context"
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/dv-net/dv-processing/pkg/avalidator"
)

// ValidateCreateAddressDTO validates the address creation request and handles business logic
func (s *Service) ValidateCreateAddressDTO(ctx context.Context, params CreateAddressDTO) error {
	// Determine address type for validation
	addressType := AddressTypeSimpleAddress
	if params.EVM {
		addressType = AddressTypeEVMAddress
	} else if params.Universal {
		addressType = AddressTypeUniversalAddress
	}

	switch addressType {
	case AddressTypeEVMAddress:
		return s.validateEVMAddress(params.Address)
	case AddressTypeUniversalAddress:
		return s.validateUniversalAddress(params)
	default:
		return s.validateSpecificAddress(ctx, params)
	}
}

func (s *Service) validateEVMAddress(address string) error {
	if !avalidator.ValidateAddressByBlockchain(address, models.BlockchainEthereum.String()) {
		return errors.New("invalid EVM address format")
	}
	return nil
}

func (s *Service) validateUniversalAddress(params CreateAddressDTO) error {
	if params.Blockchain == nil {
		return errors.New("blockchain is required for universal addresses")
	}

	if !avalidator.ValidateAddressByBlockchain(params.Address, params.Blockchain.String()) {
		return errors.New("invalid address format for the specified blockchain")
	}

	return nil
}

func (s *Service) validateSpecificAddress(ctx context.Context, params CreateAddressDTO) error {
	if params.CurrencyID == "" {
		return errors.New("currency_id is required for specific addresses")
	}

	// Get currency information for validation
	currency, err := s.currencyService.GetCurrencyByID(ctx, params.CurrencyID)
	if err != nil {
		return errors.New("invalid currency")
	}

	// Validate address based on blockchain if specified, or use currency's blockchain
	var blockchainToValidate string
	if params.Blockchain != nil {
		blockchainToValidate = params.Blockchain.String()
		// If blockchain is specified, ensure it matches the currency's blockchain for consistency
		if *params.Blockchain != *currency.Blockchain {
			// Allow EVM addresses to be used across EVM chains
			if !params.Blockchain.IsEVMLike() || !currency.Blockchain.IsEVMLike() {
				return errors.New("blockchain does not match currency blockchain")
			}
		}
	} else {
		blockchainToValidate = currency.Blockchain.String()
	}

	// Validate address format
	if !avalidator.ValidateAddressByBlockchain(params.Address, blockchainToValidate) {
		return errors.New("invalid address format for the specified blockchain")
	}

	// Check if address already exists for this user and currency
	exists, err := s.CheckAddressExists(ctx, params.UserID, params.Address, params.CurrencyID)
	if err != nil {
		return fmt.Errorf("failed to check address existence: %w", err)
	}
	if exists {
		return errors.New("address already exists for this currency")
	}

	return nil
}

// ValidateAddWithdrawalRuleRequest validates add withdrawal rule request
// func (s *Service) ValidateAddWithdrawalRuleRequest(ctx context.Context, userID uuid.UUID, id *string, address *string, blockchain *models.Blockchain, isEVM, isUniversal bool) error {
// 	switch {
// 	case isEVM:
// 		// EVM addition - validate address
// 		if address == nil {
// 			return errors.New("address is required for EVM withdrawal rule addition")
// 		}
// 	case isUniversal:
// 		// Universal addition - validate address and blockchain
// 		if address == nil {
// 			return errors.New("address is required for universal withdrawal rule addition")
// 		}
// 		if blockchain == nil {
// 			return errors.New("blockchain is required for universal withdrawal rule addition")
// 		}
// 		if err := blockchain.Valid(); err != nil {
// 			return errors.New("invalid blockchain")
// 		}
// 	default:
// 		// Individual addition - validate ID
// 		if id == nil {
// 			return errors.New("id is required for individual withdrawal rule addition")
// 		}
// 		if _, err := uuid.Parse(*id); err != nil {
// 			return errors.New("invalid address ID format")
// 		}
// 	}

// 	return nil
// }
