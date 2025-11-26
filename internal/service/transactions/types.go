package transactions

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
)

type UserTransactionsFilter struct {
	UserID      uuid.UUID
	Currencies  []string
	StoreUuids  []uuid.UUID
	ToAddress   string
	FromAddress string
	Type        *models.TransactionsType
}

type StatisticsParams struct {
	User        models.User
	Resolution  *StatisticsResolution
	DateFrom    *string
	DateTo      *string
	CurrencyIDs []string
	StoreUUIDS  []uuid.UUID
}

type StatisticsResolution string //	@name	StatisticsResolution

func (b StatisticsResolution) String() string { return string(b) }

func (b StatisticsResolution) Valid() error {
	switch b {
	case StatisticsResolutionHour,
		StatisticsResolutionDay,
		StatisticsResolutionWeek,
		StatisticsResolutionMonth,
		StatisticsResolutionQuarter,
		StatisticsResolutionYear:
		return nil
	}

	return fmt.Errorf("invalid resolution: %s", string(b))
}

const (
	StatisticsResolutionHour    StatisticsResolution = "hour"
	StatisticsResolutionDay     StatisticsResolution = "day"
	StatisticsResolutionWeek    StatisticsResolution = "week"
	StatisticsResolutionMonth   StatisticsResolution = "month"
	StatisticsResolutionQuarter StatisticsResolution = "quarter"
	StatisticsResolutionYear    StatisticsResolution = "year"
)

type StatisticsDTO struct {
	Date              time.Time                     `json:"date" format:"date-time"`
	Type              StatisticsResolution          `json:"type"`
	SumUsd            string                        `json:"sum_usd"`
	TransactionsCount int64                         `json:"transactions_count"`
	DetailsByCurrency map[string]CurrencyDetailsDTO `json:"details_by_currency"`
} //	@name	StatisticsDTO

type CurrencyDetailsDTO struct {
	TxCount int64           `json:"tx_count"`
	SumUSD  decimal.Decimal `json:"sum_usd"`
}

type TransactionInfoDto struct {
	ID               uuid.UUID
	IsConfirmed      bool
	UserID           uuid.UUID
	StoreID          *uuid.UUID
	ReceiptID        *uuid.UUID
	Wallet           TransactionWalletInfoDto
	CurrencyID       string
	Blockchain       string
	TxHash           string
	BcUniqKey        string
	Type             string
	FromAddress      string
	ToAddress        string
	Amount           decimal.Decimal
	AmountUsd        *decimal.Decimal
	Fee              decimal.Decimal
	NetworkCreatedAt *time.Time
	WebhookHistory   []TransactionWhHistoryDto
	CreatedAt        *time.Time
	UpdatedAt        *time.Time
}

type ShortTransactionInfo struct {
	IsConfirmed  bool
	CurrencyCode string
	Hash         string
	Amount       string
	AmountUSD    string
	Type         string
	CreatedAt    time.Time
}

type TransactionWalletInfoDto struct {
	ID              uuid.UUID
	WalletStoreID   uuid.UUID
	StoreExternalID string
	WalletCreatedAt time.Time
	WalletUpdatedAt *time.Time
}

type TransactionWhHistoryDto struct {
	ID                 uuid.UUID
	StoreID            uuid.UUID
	WhType             string
	URL                string
	WhStatus           string
	Request            []byte
	Response           *string
	ResponseStatusCode int
	CreatedAt          *time.Time
}

type WalletWithTransactionsInfo struct {
	Address         string                  `json:"address"`
	StoreUUID       uuid.UUID               `json:"store_uuid"`
	WalletID        uuid.UUID               `json:"wallet_id"`
	StoreExternalID string                  `json:"store_external_id"`
	Currencies      []string                `json:"currencies"`
	Transactions    []WalletTransactionInfo `json:"transactions"`
} //	@name	WalletWithTransactionsInfo

type WalletTransactionInfo struct {
	CurrencyID string    `json:"currency_id"`
	Hash       string    `json:"hash"`
	From       string    `json:"from"`
	To         string    `json:"to"`
	CreatedAt  time.Time `json:"created_at"`
} //	@name	WalletTransactionInfo

type UserTransactionModel struct {
	StoreName        string    `json:"store_name" csv:"store_name" excel:"store_name"`
	ReceiptID        string    `json:"receipt_id" csv:"receipt_id" excel:"receipt_id"`
	CurrencyID       string    `json:"currency_id" csv:"currency_id" excel:"currency_id"`
	Blockchain       string    `json:"blockchain" csv:"blockchain" excel:"blockchain"`
	Name             string    `json:"name" csv:"name" excel:"name"`
	TxHash           string    `json:"tx_hash" csv:"tx_hash" excel:"tx_hash"`
	Type             string    `json:"type" csv:"type" excel:"type"`
	FromAddress      string    `json:"from_address" csv:"from_address" excel:"from_address"`
	ToAddress        string    `json:"to_address" csv:"to_address" excel:"to_address"`
	Amount           string    `json:"amount" csv:"amount" excel:"amount"`
	AmountUsd        string    `json:"amount_usd" csv:"amount_usd" excel:"amount_usd"`
	Fee              string    `json:"fee" csv:"fee" excel:"fee"`
	NetworkCreatedAt time.Time `json:"network_created_at" csv:"network_created_at" excel:"network_created_at"`
	CreatedAt        time.Time `json:"created_at" csv:"created_at" excel:"created_at"`
	IsSystem         bool      `json:"is_system" csv:"is_system" excel:"is_system"`
	UserEmail        *string   `json:"user_email" csv:"user_email" excel:"user_email"`
}
