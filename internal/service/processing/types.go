package processing

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallet_addresses"

	"github.com/dv-net/dv-merchant/internal/models"

	commonv1 "github.com/dv-net/dv-processing/api/processing/common/v1"
	"github.com/google/uuid"
)

type OwnerHotWalletParams struct {
	OwnerID             uuid.UUID
	Blockchain          models.Blockchain
	CustomerID          string
	BitcoinAddressType  *BitcoinAddressType
	LitecoinAddressType *LitecoinAddressType
}

type BitcoinAddressType int32
type LitecoinAddressType int32

func (o BitcoinAddressType) ToPb() commonv1.BitcoinAddressType {
	return commonv1.BitcoinAddressType(o)
}
func (o LitecoinAddressType) ToPb() commonv1.LitecoinAddressType {
	return commonv1.LitecoinAddressType(o)
}

const (
	BitcoinAddressTypeUnspecified BitcoinAddressType = iota
	BitcoinAddressTypeP2pkh                          // Legacy
	BitcoinAddressTypeP2sh                           // SegWit
	BitcoinAddressTypeP2wpkh                         // Native SegWit or Bech32
	BitcoinAddressTypeP2tr                           // Taproot address or Bech32m
)

const (
	LitecoinAddressTypeUnspecified LitecoinAddressType = iota
	LitecoinAddressTypeP2pkh                           // Legacy
	LitecoinAddressTypeP2sh                            // SegWit
	LitecoinAddressTypeP2wpkh                          // Native SegWit or Bech32
	LitecoinAddressTypeP2tr                            // Taproot address or Bech32m
)

func ConvertToBitcoinAddressType(addressType string) BitcoinAddressType {
	switch addressType {
	case BitcoinAddressTypeP2pkhString:
		return BitcoinAddressTypeP2pkh
	case BitcoinAddressTypeP2shString:
		return BitcoinAddressTypeP2sh
	case BitcoinAddressTypeP2wpkhString:
		return BitcoinAddressTypeP2wpkh
	case BitcoinAddressTypeP2trString:
		return BitcoinAddressTypeP2tr
	default:
		return BitcoinAddressTypeUnspecified
	}
}

func ConvertToLitecoinAddressType(addressType string) LitecoinAddressType {
	switch addressType {
	case LitecoinAddressTypeP2pkhString:
		return LitecoinAddressTypeP2pkh
	case LitecoinAddressTypeP2shString:
		return LitecoinAddressTypeP2sh
	case LitecoinAddressTypeP2wpkhString:
		return LitecoinAddressTypeP2wpkh
	case LitecoinAddressTypeP2trString:
		return LitecoinAddressTypeP2tr
	default:
		return LitecoinAddressTypeUnspecified
	}
}

const (
	BitcoinAddressTypeP2pkhString  = "P2PKH"
	BitcoinAddressTypeP2shString   = "P2SH"
	BitcoinAddressTypeP2wpkhString = "P2WPKH"
	BitcoinAddressTypeP2trString   = "P2TR"

	LitecoinAddressTypeP2pkhString  = "P2PKH"
	LitecoinAddressTypeP2shString   = "P2SH"
	LitecoinAddressTypeP2wpkhString = "P2WPKH"
	LitecoinAddressTypeP2trString   = "P2TR"
)

type CreateOwnerHotWalletParams struct {
	OwnerID             uuid.UUID
	Blockchain          models.Blockchain
	CustomerID          string
	BitcoinAddressType  *BitcoinAddressType
	LitecoinAddressType *LitecoinAddressType
}

type AttachOwnerColdWalletsParams struct {
	OwnerID    uuid.UUID
	Blockchain models.Blockchain
	Addresses  []string
	TOTP       string
}

type GetOwnerColdWalletsParams struct {
	OwnerID    uuid.UUID
	Blockchain models.Blockchain
}

type GetOwnerProcessingWalletsParams struct {
	OwnerID    uuid.UUID
	Blockchain *models.Blockchain
	Tiny       *bool
}

type HotWallet struct {
	Address string
}

type BlockchainWalletsList struct {
	Wallets    string            `json:"wallets"`
	Blockchain models.Blockchain `json:"blockchain"`
}

type OwnerWalletTransactionsParams struct {
	OwnerID    uuid.UUID
	Address    string
	Blockchain models.Blockchain
}

type FundsWithdrawalParams struct {
	OwnerID            uuid.UUID
	RequestID          uuid.UUID
	Blockchain         models.Blockchain
	FromAddress        []string
	ToAddress          []string
	ContractAddress    string
	WholeAmount        bool
	Amount             string
	Fee                *uint64
	Threshold          *uint64
	Immediately        *bool
	DryRun             *bool
	IncomingWalletType WalletType
	Kind               *string
}

type WalletType string

type WithdrawalStatus int

const (
	UnknownWithdrawalStatus WithdrawalStatus = iota
	PendingWithdrawalStatus
	AcceptedWithdrawalStatus
	SuccessWithdrawalStatus
	FailedWithdrawalStatus
)

type FundsWithdrawalResult struct {
	WithdrawalStatus WithdrawalStatus
	TxHash           *string
	Message          *string
}

type TransactionDto struct {
	BlockHeight     uint64
	Hash            string
	CreatedAt       *time.Time
	AddressFrom     string
	AddressTo       string
	Amount          string
	TxType          string
	Confirmations   uint64
	Fee             *string
	ContractAddress *string
	Result          *string
	Order           *string
}

type Settings struct {
	BaseURL             string `json:"base_url"`
	ProcessingClientID  string `json:"processing_client_id"`
	ProcessingClientKey string `json:"processing_client_key"`
} // @name ProcessingRootSettings

type TwoFactorAuthData struct {
	Secret      string `json:"secret,omitempty"`
	IsConfirmed bool   `json:"is_confirmed"`
} // @name TwoFactorAuthData

type GetOwnerPrivateKeysData struct {
	Keys map[string]*KeyPairSequence `json:"keys"`
} // @name GetOwnerPrivateKeysData

type GetOwnerHotWalletKeysData struct {
	Entries            []HotWalletKeyPair                                     `json:"entries"`
	AllSelectedWallets []*repo_wallet_addresses.FilterOwnerWalletAddressesRow `json:"-"`
} // @name GetOwnerHotWalletKeysData

type HotWalletKeyPair struct {
	Name  string       `json:"name"`
	Items []HotKeyPair `json:"items"`
}

type HotKeyPair struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Address    string `json:"address"`
}

type KeyPairSequence struct {
	Pairs []KeyPair `json:"pairs"`
} // @name KeyPairSequence

type KeyPair struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	Address    string `json:"address"`
	Kind       string `json:"kind"`
} // @name KeyPair

type OwnerSeedData struct {
	Mnemonic   string `json:"mnemonic"`
	PassPhrase string `json:"pass_phrase"`
} // @name OwnerSeedData

type Asset struct {
	Identity  string `json:"identity"`
	Amount    string `json:"amount"`
	AmountUSD string `json:"amount_usd"`
} // @name Asset

type BlockchainAdditionalData struct {
	TronData *TronData `json:"tron_data"`
} // @name BlockchainAdditionalData

type TronData struct {
	AvailableEnergyForUse    string `json:"available_energy_for_use"`
	TotalEnergy              string `json:"total_energy"`
	AvailableBandwidthForUse string `json:"available_bandwidth_for_use"`
	TotalBandwidth           string `json:"total_bandwidth"`
	StackedTrx               string `json:"stacked_trx"`
	StackedEnergy            string `json:"stacked_energy"`
	StackedEnergyTrx         string `json:"stacked_energy_trx"`
	StackedBandwidth         string `json:"stacked_bandwidth"`
	StackedBandwidthTrx      string `json:"stacked_bandwidth_trx"`
	TotalUsedEnergy          string `json:"total_used_energy"`
	TotalUsedBandwidth       string `json:"total_used_bandwidth"`
} // @name TronData

type WalletProcessing struct {
	Address        string                    `json:"address"`
	Blockchain     models.Blockchain         `json:"blockchain"`
	Assets         []*Asset                  `json:"assets,omitempty"`
	AdditionalData *BlockchainAdditionalData `json:"additional_data,omitempty"`
} // @name WalletProcessing

type Info struct {
	Version string
	Hash    string
}

type NewVersion struct {
	Name             string
	AvailableVersion string
	InstalledVersion string
	NeedForUpdate    bool
} // @name NewVersion

type RegisterOwnerInfo struct {
	OwnerID uuid.UUID `json:"owner_id"`
} // @name RegisterOwnerInfo

type GetOwnerHotWalletKeysParams struct {
	WalletAddressIDs           []uuid.UUID
	ExcludedWalletAddressesIDs []uuid.UUID
}
