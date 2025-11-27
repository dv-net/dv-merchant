package processing_request

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ProcessingWebhook struct {
	Blockchain       models.Blockchain        `json:"blockchain"`
	Hash             string                   `json:"hash"`
	RequestID        uuid.NullUUID            `json:"request_id"`
	NetworkCreatedAt time.Time                `json:"network_created_at"`
	FromAddress      string                   `json:"from_address,omitempty"`
	ToAddress        string                   `json:"to_address,omitempty"`
	Amount           string                   `json:"amount"`
	ContractAddress  string                   `json:"contract_address,omitempty"`
	Status           models.TransactionStatus `json:"status"`
	IsSystem         bool                     `json:"is_system"`
	Fee              string                   `json:"fee"`
	Confirmations    uint64                   `json:"confirmations"`
	TxUniqKey        string                   `json:"tx_uniq_key,omitempty"`
	ExternalWalletID string                   `json:"external_wallet_id,omitempty"`
	WalletType       models.WalletType        `json:"wallet_type"`
	Kind             models.WebhookKind       `json:"kind"`
} //	@name	ProcessingWebhook

type TransferStatusWebhook struct {
	Kind               models.WebhookKind           `json:"kind"`
	Status             models.TransferStatus        `json:"status" `
	SystemTransactions []TransferSystemTransactions `json:"system_transactions" validate:"dive"`
	Step               string                       `json:"step"`
	ErrorMessage       *string                      `json:"error_message,omitempty"`
	RequestID          *uuid.UUID                   `json:"request_id,omitempty" format:"uuid"`
} //	@name	TransferStatusWebhook

type TransferSystemTransactions struct {
	ID                uuid.UUID                         `json:"id"`
	TransferID        uuid.UUID                         `json:"transfer_id"`
	TxHash            string                            `json:"tx_hash"`
	BandwidthAmount   decimal.Decimal                   `json:"bandwidth_amount" validate:"required,decimal_gte=0"`
	EnergyAmount      decimal.Decimal                   `json:"energy_amount" validate:"required,decimal_gte=0"`
	NativeTokenAmount decimal.Decimal                   `json:"native_token_amount" validate:"required,decimal_gte=0"`
	NativeTokenFee    decimal.Decimal                   `json:"native_token_fee" validate:"required,decimal_gte=0"`
	TxType            models.TransferTransactionType    `json:"tx_type" validate:"required,oneof=resource_delegation resource_reclaim send_burn_base_asset account_activation transfer"`
	Status            models.TransferTransactionsStatus `json:"status" validate:"required,oneof=pending unconfirmed confirmed failed"`
	Step              string                            `json:"step"`
} //	@name	TransferSystemTransactions
