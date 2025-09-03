package models

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type TransactionStatus string // @name TransactionStatus

func (o TransactionStatus) String() string { return string(o) }

const (
	TransactionStatusWaitingConfirmations TransactionStatus = "waiting_confirmations"
	TransactionStatusInMempool            TransactionStatus = "in_mempool"
	TransactionStatusCompleted            TransactionStatus = "completed"
	TransactionStatusFailed               TransactionStatus = "failed"
)

type TransactionsType string // @name TransactionsType

func (t TransactionsType) String() string { return string(t) }

const (
	TransactionsTypeTransfer                 TransactionsType = "transfer"
	TransactionsTypeDeposit                  TransactionsType = "deposit"
	TransactionsTypeWithdrawalFromProcessing TransactionsType = "withdrawal_from_processing"
)

func (t TransactionsType) RequiresWebhookToStore() bool {
	switch t {
	case TransactionsTypeDeposit:
		// Only deposit transaction requires
		return true
	default:
		return false
	}
}

type ITransaction interface { //nolint:interfacebloat
	GetID() uuid.UUID
	GetTxHash() string
	GetBcUniqKey() *string
	GetType() TransactionsType
	GetCreatedAt() pgtype.Timestamp
	GetNetworkCreatedAt() pgtype.Timestamp
	GetAmount() decimal.Decimal
	GetAmountUsd() decimal.Decimal
	GetWalletID() uuid.NullUUID
	GetStoreID() uuid.UUID
	GetCurrencyID() string
	IsConfirmed() bool
}

// transaction

func (tx Transaction) GetID() uuid.UUID {
	return tx.ID
}

func (tx Transaction) GetStoreID() uuid.UUID {
	return tx.StoreID.UUID
}

func (tx Transaction) GetTxHash() string {
	return tx.TxHash
}

func (tx Transaction) GetBcUniqKey() *string {
	return tx.BcUniqKey
}

func (tx Transaction) GetType() TransactionsType {
	return tx.Type
}

func (tx Transaction) GetCreatedAt() pgtype.Timestamp {
	return tx.CreatedAt
}

func (tx Transaction) GetNetworkCreatedAt() pgtype.Timestamp {
	return tx.NetworkCreatedAt
}

func (tx Transaction) GetAmount() decimal.Decimal {
	return tx.Amount
}

func (tx Transaction) GetAmountUsd() decimal.Decimal {
	return tx.AmountUsd.Decimal
}

func (tx Transaction) GetWalletID() uuid.NullUUID {
	return tx.WalletID
}

func (tx Transaction) IsConfirmed() bool {
	return true
}

func (tx Transaction) GetCurrencyID() string {
	return tx.CurrencyID
}

// Unconfirmed transaction

func (utx UnconfirmedTransaction) GetID() uuid.UUID {
	return utx.ID
}

func (utx UnconfirmedTransaction) GetTxHash() string {
	return utx.TxHash
}

func (utx UnconfirmedTransaction) GetBcUniqKey() *string {
	return utx.BcUniqKey
}

func (utx UnconfirmedTransaction) GetType() TransactionsType {
	return utx.Type
}

func (utx UnconfirmedTransaction) GetCreatedAt() pgtype.Timestamp {
	return utx.CreatedAt
}

func (utx UnconfirmedTransaction) GetNetworkCreatedAt() pgtype.Timestamp {
	return utx.NetworkCreatedAt
}

func (utx UnconfirmedTransaction) GetAmount() decimal.Decimal {
	return utx.Amount
}

func (utx UnconfirmedTransaction) GetAmountUsd() decimal.Decimal {
	return utx.AmountUsd.Decimal
}

func (utx UnconfirmedTransaction) GetWalletID() uuid.NullUUID {
	return utx.WalletID
}

func (utx UnconfirmedTransaction) GetStoreID() uuid.UUID {
	return utx.StoreID.UUID
}

func (utx UnconfirmedTransaction) IsConfirmed() bool {
	return false
}

func (utx UnconfirmedTransaction) GetCurrencyID() string {
	return utx.CurrencyID
}
