package callback

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
)

type WebhookDtoInterface interface {
	GetToAddress() string
	GetAddressForUpdateBalance() string
	GetStatus() models.TransactionStatus
	GetCurrency() *models.Currency
}

func (dto DepositWebhookDto) GetToAddress() string {
	return dto.ToAddress
}

func (dto DepositWebhookDto) GetAddressForUpdateBalance() string {
	return dto.ToAddress
}

func (dto DepositWebhookDto) GetCurrency() *models.Currency {
	return dto.Currency
}

func (dto DepositWebhookDto) GetStatus() models.TransactionStatus {
	return dto.Status
}

func (dto TransferWebhookDto) GetToAddress() string {
	return dto.ToAddress
}

func (dto TransferWebhookDto) GetAddressForUpdateBalance() string {
	return dto.FromAddress
}

func (dto TransferWebhookDto) GetCurrency() *models.Currency {
	return dto.Currency
}

func (dto TransferWebhookDto) GetStatus() models.TransactionStatus {
	return dto.Status
}

type ProcessingWebhook struct {
	Blockchain       models.Blockchain        `json:"blockchain"`
	Hash             string                   `json:"hash"`
	NetworkCreatedAt time.Time                `json:"network_created_at"`
	FromAddress      string                   `json:"from_address"`
	ToAddress        string                   `json:"to_address"`
	Amount           string                   `json:"amount"`
	Fee              string                   `json:"fee"`
	ContractAddress  string                   `json:"contract_address"`
	Status           models.TransactionStatus `json:"status"`
	Confirmations    uint64                   `json:"confirmations"`
	TransactionType  models.TransactionsType  `json:"transaction_type"`
	TxUniqKey        string                   `json:"tx_uniq_key,omitempty"`
	ExternalWalletID uuid.UUID                `json:"external_wallet_id,omitempty"`
}

type DepositWebhookDto struct {
	Blockchain       models.Blockchain        `json:"blockchain"`
	Hash             string                   `json:"hash"`
	NetworkCreatedAt time.Time                `json:"network_created_at"`
	FromAddress      string                   `json:"from_address"`
	ToAddress        string                   `json:"to_address"`
	Amount           string                   `json:"amount"`
	Fee              string                   `json:"fee"`
	ContractAddress  string                   `json:"contract_address"`
	Status           models.TransactionStatus `json:"status"`
	IsSystem         bool                     `json:"is_system"`
	Confirmations    uint64                   `json:"confirmations"`
	WalletType       models.WalletType        `json:"wallet_type"`
	TxUniqKey        string                   `json:"tx_uniq_key,omitempty"`
	ExternalWalletID *uuid.UUID               `json:"external_wallet_id,omitempty"`
	Currency         *models.Currency
}

type TransferWebhookDto struct {
	Blockchain       models.Blockchain        `json:"blockchain"`
	TransferID       uuid.NullUUID            `json:"transfer_id"`
	Hash             string                   `json:"hash"`
	NetworkCreatedAt time.Time                `json:"network_created_at"`
	FromAddress      string                   `json:"from_address"`
	ToAddress        string                   `json:"to_address"`
	Amount           string                   `json:"amount"`
	Fee              string                   `json:"fee"`
	ContractAddress  string                   `json:"contract_address"`
	Status           models.TransactionStatus `json:"status"`
	IsSystem         bool                     `json:"is_system"`
	Confirmations    uint64                   `json:"confirmations"`
	TxUniqKey        string                   `json:"tx_uniq_key,omitempty"`
	WalletType       models.WalletType        `json:"wallet_type"`
	ExternalWalletID *uuid.UUID               `json:"external_wallet_id,omitempty"`
	Currency         *models.Currency
}
