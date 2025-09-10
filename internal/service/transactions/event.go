package transactions

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
)

type TransactionEvent interface { //nolint:interfacebloat
	EventType() string
	GetStore() models.Store
	GetTx() models.ITransaction
	GetCurrency() models.Currency
	GetWebhookEvent() models.WebhookEvent
	GetStoreExternalID() string
	GetDatabaseTx() pgx.Tx
	GetWalletLocale() string
	GetWalletEmail() string
	GetUsdFee() decimal.Decimal
	GetExchangeRate() decimal.Decimal
}

const (
	DepositReceivedEventType                  = "deposit_received"
	DepositUnconfirmedEventType               = "deposit_unconfirmed"
	DepositReceiptSentEventType               = "deposit_receipt_sent"
	WithdrawalFromProcessingReceivedEventType = "withdrawal_from_processing_received"
)

type DepositReceivedEvent struct {
	Tx              models.Transaction
	Store           models.Store
	Currency        models.Currency
	StoreExternalID string
	WebhookEvent    models.WebhookEvent
	DBTx            pgx.Tx
}

type DepositUnconfirmedEvent struct {
	Tx              models.UnconfirmedTransaction
	Store           models.Store
	Currency        models.Currency
	StoreExternalID string
	WebhookEvent    models.WebhookEvent
	DBTx            pgx.Tx
}

type WithdrawalFromProcessingReceivedEvent struct {
	WithdrawalID string
	Tx           models.Transaction
	Store        models.Store
	Currency     models.Currency
	WebhookEvent models.WebhookEvent
	DBTx         pgx.Tx
}

type DepositReceiptSentEvent struct {
	ReceiptID       string
	Tx              models.Transaction
	Store           models.Store
	Currency        models.Currency
	StoreExternalID string
	WebhookEvent    models.WebhookEvent
	ExchangeRate    decimal.Decimal
	WalletLocale    string
	WalletEmail     string
	UsdFee          decimal.Decimal
	DBTx            pgx.Tx
}

// withdrawal_from_processing received event

func (e WithdrawalFromProcessingReceivedEvent) Type() event.Type {
	return WithdrawalFromProcessingReceivedEventType
}

func (e WithdrawalFromProcessingReceivedEvent) String() string {
	return fmt.Sprintf("DepositReceived: tx=%v, store=%d, webhook_event=%v", e.Tx, e.Store.ID, e.WebhookEvent)
}

func (e WithdrawalFromProcessingReceivedEvent) EventType() string {
	return WithdrawalFromProcessingReceivedEventType
}

func (e WithdrawalFromProcessingReceivedEvent) GetStore() models.Store {
	return e.Store
}

func (e WithdrawalFromProcessingReceivedEvent) GetTx() models.ITransaction {
	return e.Tx
}

func (e WithdrawalFromProcessingReceivedEvent) GetCurrency() models.Currency {
	return e.Currency
}

func (e WithdrawalFromProcessingReceivedEvent) GetWebhookEvent() models.WebhookEvent {
	return e.WebhookEvent
}

func (e WithdrawalFromProcessingReceivedEvent) GetStoreExternalID() string {
	return ""
}

func (e WithdrawalFromProcessingReceivedEvent) Locale() string {
	return ""
}

func (e WithdrawalFromProcessingReceivedEvent) GetDatabaseTx() pgx.Tx {
	return e.DBTx
}

func (e WithdrawalFromProcessingReceivedEvent) GetWalletLocale() string {
	return ""
}
func (e WithdrawalFromProcessingReceivedEvent) GetWalletEmail() string {
	return ""
}

func (e WithdrawalFromProcessingReceivedEvent) GetUsdFee() decimal.Decimal {
	return decimal.Zero
}

func (e WithdrawalFromProcessingReceivedEvent) GetExchangeRate() decimal.Decimal {
	return decimal.Zero
}

// deposit ReceivedEvent

func (e DepositReceivedEvent) Type() event.Type {
	return DepositReceivedEventType
}

func (e DepositReceivedEvent) EventType() string {
	return DepositReceivedEventType
}

func (e DepositReceivedEvent) GetStore() models.Store {
	return e.Store
}

func (e DepositReceivedEvent) GetTx() models.ITransaction {
	return e.Tx
}

func (e DepositReceivedEvent) GetCurrency() models.Currency {
	return e.Currency
}

func (e DepositReceivedEvent) GetWebhookEvent() models.WebhookEvent {
	return e.WebhookEvent
}

func (e DepositReceivedEvent) GetStoreExternalID() string {
	return e.StoreExternalID
}

func (e DepositReceivedEvent) GetDatabaseTx() pgx.Tx {
	return e.DBTx
}

func (e DepositReceivedEvent) String() string {
	return fmt.Sprintf("DepositReceived: tx=%v, store=%d, store_external_id=%s, webhook_event=%v", e.Tx, e.Store.ID, e.StoreExternalID, e.WebhookEvent)
}

func (e DepositReceivedEvent) GetWalletLocale() string {
	return ""
}
func (e DepositReceivedEvent) GetWalletEmail() string {
	return ""
}

func (e DepositReceivedEvent) GetUsdFee() decimal.Decimal {
	return decimal.Zero
}

func (e DepositReceivedEvent) GetExchangeRate() decimal.Decimal {
	return decimal.Zero
}

// deposit unconfirmed event

func (e DepositUnconfirmedEvent) Type() event.Type {
	return DepositUnconfirmedEventType
}

func (e DepositUnconfirmedEvent) EventType() string {
	return DepositUnconfirmedEventType
}

func (e DepositUnconfirmedEvent) GetStore() models.Store {
	return e.Store
}

func (e DepositUnconfirmedEvent) GetTx() models.ITransaction {
	return e.Tx
}

func (e DepositUnconfirmedEvent) GetCurrency() models.Currency {
	return e.Currency
}

func (e DepositUnconfirmedEvent) GetWebhookEvent() models.WebhookEvent {
	return e.WebhookEvent
}

func (e DepositUnconfirmedEvent) GetStoreExternalID() string {
	return e.StoreExternalID
}

func (e DepositUnconfirmedEvent) GetDatabaseTx() pgx.Tx {
	return e.DBTx
}

func (e DepositUnconfirmedEvent) String() string {
	return fmt.Sprintf("DepositReceived: tx=%v, store=%d, store_external_id=%s, webhook_event=%v", e.Tx, e.Store.ID, e.StoreExternalID, e.WebhookEvent)
}

func (e DepositUnconfirmedEvent) GetWalletLocale() string {
	return ""
}
func (e DepositUnconfirmedEvent) GetWalletEmail() string {
	return ""
}

func (e DepositUnconfirmedEvent) GetUsdFee() decimal.Decimal {
	return decimal.Zero
}

func (e DepositUnconfirmedEvent) GetExchangeRate() decimal.Decimal {
	return decimal.Zero
}

// deposit receipt

func (e DepositReceiptSentEvent) Type() event.Type {
	return DepositReceiptSentEventType
}

func (e DepositReceiptSentEvent) EventType() string {
	return DepositReceiptSentEventType
}

func (e DepositReceiptSentEvent) GetStore() models.Store {
	return e.Store
}

func (e DepositReceiptSentEvent) GetTx() models.ITransaction {
	return e.Tx
}

func (e DepositReceiptSentEvent) GetCurrency() models.Currency {
	return e.Currency
}

func (e DepositReceiptSentEvent) GetWebhookEvent() models.WebhookEvent {
	return e.WebhookEvent
}

func (e DepositReceiptSentEvent) GetStoreExternalID() string {
	return e.StoreExternalID
}

func (e DepositReceiptSentEvent) GetDatabaseTx() pgx.Tx {
	return e.DBTx
}

func (e DepositReceiptSentEvent) GetWalletLocale() string {
	return e.WalletLocale
}

func (e DepositReceiptSentEvent) GetWalletEmail() string {
	return e.WalletEmail
}

func (e DepositReceiptSentEvent) GetUsdFee() decimal.Decimal {
	return e.UsdFee
}

func (e DepositReceiptSentEvent) GetExchangeRate() decimal.Decimal {
	return e.ExchangeRate
}

func (e DepositReceiptSentEvent) String() string {
	return fmt.Sprintf(
		"DepositReceiptSent: receipt_id=%s, tx=%v, store=%d, user_email=%s, store_external_id=%s, webhook_event=%v",
		e.ReceiptID, e.Tx, e.Store.ID, e.WalletEmail, e.StoreExternalID, e.WebhookEvent,
	)
}
