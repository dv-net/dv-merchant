package wallet

import (
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallets"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type LoadPrivateKeyDTO struct {
	User                       *models.User
	Otp                        string
	WalletAddressIDs           []uuid.UUID
	ExcludedWalletAddressesIDs []uuid.UUID
	FileType                   string
	IP                         string
}

type AddressesTotalBalance struct {
	TotalUSD  decimal.Decimal `json:"total_usd"`
	TotalDust decimal.Decimal `json:"total_dust"`
} //	@name	AddressesTotalBalance

type CalcBalanceDTO struct {
	Address    string
	Blockchain *models.Blockchain
}

type ConvertAddressDTO struct {
	Address       *string
	LegacyAddress *string
}

type GetProcessingWalletsDTO struct {
	OwnerID     uuid.UUID           `json:"owner_id,omitempty"`
	Blockchains []models.Blockchain `json:"blockchains"`
	Currencies  []string            `json:"currencies"`
}

type FetchTronStatsResolution string

const (
	FetchTronStatsResolutionHour  = "hour"
	FetchTronStatsResolutionDay   = "day"
	FetchTronStatsResolutionMonth = "month"
)

type FetchTronStatisticsParams struct {
	DateFrom   *string
	DateTo     *string
	Resolution string
}

type CombinedStats struct {
	StakedBandwidth    decimal.Decimal `json:"staked_bandwidth"`
	StakedEnergy       decimal.Decimal `json:"staked_energy"`
	DelegatedEnergy    decimal.Decimal `json:"delegated_energy"`
	DelegatedBandwidth decimal.Decimal `json:"delegated_bandwidth"`
	AvailableBandwidth decimal.Decimal `json:"available_bandwidth"`
	AvailableEnergy    decimal.Decimal `json:"available_energy"`
	TransferCount      int64           `json:"transfer_count"`
	TotalTrxFee        decimal.Decimal `json:"total_trx_fee"`
	TotalBandwidthUsed decimal.Decimal `json:"total_bandwidth_used"`
	TotalEnergyUsed    decimal.Decimal `json:"total_energy_used"`
} //	@name	CombinedStats

type CreateStoreWalletWithAddressDTO struct {
	StoreID         uuid.UUID `db:"store_id" json:"store_id"`
	StoreExternalID string    `db:"store_external_id" json:"store_external_id"`
	Email           *string   `db:"email" json:"email"`
	IP              *string   `db:"ip" json:"ip"`
	UntrustedEmail  *string   `db:"untrusted_email" json:"untrusted_email"`
	Locale          *string   `db:"locale" json:"locale"`
}

func (dto *CreateStoreWalletWithAddressDTO) ToCreateParams() repo_wallets.CreateParams {
	params := repo_wallets.CreateParams{
		StoreID:         dto.StoreID,
		StoreExternalID: dto.StoreExternalID,
		Email:           pgtypeutils.EncodeText(dto.Email),
		IpAddress:       pgtypeutils.EncodeText(dto.IP),
		UntrustedEmail:  pgtypeutils.EncodeText(dto.UntrustedEmail),
	}

	if dto.Locale != nil && *dto.Locale != "" {
		params.Locale = *dto.Locale
	}
	return params
}
