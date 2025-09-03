package withdrawal_requests

import (
	"testing"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/util"
)

func Test_CreateAddressBook(t *testing.T) {
	tests := []struct {
		name      string
		request   CreateAddressBookRequest
		wantError bool
	}{
		// Simple address creation tests
		{
			name: "SimpleAddressValid",
			request: CreateAddressBookRequest{
				Address:     "0xf98048Fa59B15efA1EC392bCE24bf8eA6d8aA422",
				CurrencyID:  "USDC.Polygon",
				IsUniversal: false,
				IsEVM:       false,
				TOTP:        "123456",
			},
			wantError: false,
		},
		{
			name: "SimpleAddressInvalidMissingCurrency",
			request: CreateAddressBookRequest{
				Address:     "0xf98048Fa59B15efA1EC392bCE24bf8eA6d8aA422",
				CurrencyID:  "",
				IsUniversal: false,
				IsEVM:       false,
				TOTP:        "123456",
			},
			wantError: true,
		},
		// EVM address creation tests
		{
			name: "EVMAddressValid",
			request: CreateAddressBookRequest{
				Address:     "0xf98048Fa59B15efA1EC392bCE24bf8eA6d8aA422",
				IsUniversal: false,
				IsEVM:       true,
				TOTP:        "123456",
			},
			wantError: false,
		},
		{
			name: "EVMAddressInvalidBothTypes",
			request: CreateAddressBookRequest{
				Address:     "0xf98048Fa59B15efA1EC392bCE24bf8eA6d8aA422",
				IsUniversal: true,
				IsEVM:       true,
				TOTP:        "123456",
			},
			wantError: true,
		},
		// Universal address creation tests
		{
			name: "UniversalAddressValid",
			request: CreateAddressBookRequest{
				Address:     "0xf98048Fa59B15efA1EC392bCE24bf8eA6d8aA422",
				IsUniversal: true,
				IsEVM:       false,
				Blockchain:  util.Pointer(models.BlockchainEthereum),
				TOTP:        "123456",
			},
			wantError: false,
		},
		{
			name: "UniversalAddressInvalidMissingBlockchain",
			request: CreateAddressBookRequest{
				Address:     "0xf98048Fa59B15efA1EC392bCE24bf8eA6d8aA422",
				IsUniversal: true,
				IsEVM:       false,
				TOTP:        "123456",
			},
			wantError: true,
		},
		{
			name: "UniversalAddressInvalidBothTypes",
			request: CreateAddressBookRequest{
				Address:     "0xf98048Fa59B15efA1EC392bCE24bf8eA6d8aA422",
				IsUniversal: true,
				IsEVM:       true,
				Blockchain:  util.Pointer(models.BlockchainEthereum),
				TOTP:        "123456",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantError && err == nil {
				t.Errorf("Expected validation error for %s, but got none", tt.name)
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected validation error for %s: %v", tt.name, err)
			}
		})
	}
}
