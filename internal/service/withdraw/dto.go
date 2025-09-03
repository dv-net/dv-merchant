package withdraw

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateWithdrawalFromProcessingDTO struct {
	CurrencyID string          `json:"currency_id"`
	Amount     decimal.Decimal `json:"amount"`
	AddressTo  string          `json:"address_to"`
	UserID     uuid.UUID       `json:"user_id"`
	StoreID    *uuid.UUID      `json:"store_id"`
	RequestID  *string         `json:"request_id"`
} //	@name	CreateWithdrawalFromProcessingDTO

type WithdrawalFromProcessingDto struct {
	Transfer    *ShortTransferDto `json:"transfer"`
	TXHash      string            `json:"tx_hash"`
	StoreID     uuid.UUID         `json:"store_id"`
	CurrencyID  string            `json:"currency_id"`
	AddressFrom string            `json:"address_from"`
	AddressTo   string            `json:"address_to"`
	Amount      decimal.Decimal   `json:"amount"`
	AmountUSD   decimal.Decimal   `json:"amount_usd"`
	CreatedAt   *time.Time        `json:"created_at"`
} //	@name	WithdrawalFromProcessingDto

type ShortTransferDto struct {
	Kind    models.TransferKind   `json:"kind"`
	Stage   models.TransferStage  `json:"stage"`
	Status  models.TransferStatus `json:"status"`
	Message *string               `json:"message"`
} //	@name	ShortTransferDto

type TransferDto struct {
	ID            uuid.UUID           `json:"id"`
	UserID        uuid.UUID           `json:"user_id"`
	OwnerID       uuid.UUID           `json:"owner_id"`
	Kind          models.TransferKind `json:"kind"`
	FromAddresses []string            `json:"from_addresses"`
	ToAddress     string              `json:"to_address"`
	Contract      string              `json:"contract"`
	Amount        decimal.Decimal     `json:"amount"`
	AmountUsd     decimal.Decimal     `json:"amount_usd"`
	CurrencyID    string              `json:"currency_id"`
	Blockchain    models.Blockchain   `json:"blockchain"`
}

type WithdrawalToProcessingDTO struct {
	WalletAddressIDs           []uuid.UUID `json:"wallet_address_ids"`            //nolint:tagliatelle
	ExcludedWalletAddressesIDs []uuid.UUID `json:"excluded_wallet_addresses_ids"` //nolint:tagliatelle
	CurrencyID                 string      `json:"currency_id"`
}

type MultipleWithdrawalDTO struct {
	WithdrawalWalletID         uuid.UUID   `json:"withdrawal_wallet_id"`
	WalletAddressIDs           []uuid.UUID `json:"wallet_address_ids"`            //nolint:tagliatelle
	ExcludedWalletAddressesIDs []uuid.UUID `json:"excluded_wallet_addresses_ids"` //nolint:tagliatelle
	CurrencyID                 string      `json:"currency_id"`
}
