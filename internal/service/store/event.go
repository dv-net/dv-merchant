package store

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
)

type TransactionEvent interface {
	EventType() string
	GetStore() models.Store
	GetTx() models.ITransaction
	GetCurrency() models.Currency
	GetWebhookEvent() models.WebhookEvent
	GetStoreExternalID() string
	GetDatabaseTx() pgx.Tx
}

const (
	DepositReceivedEventType                  = "deposit_received"
	DepositUnconfirmedEventType               = "deposit_unconfirmed"
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
	WithdrawalID string `json:"withdrawal_id"`
	Tx           models.Transaction
	Store        models.Store
	Currency     models.Currency
	WebhookEvent models.WebhookEvent
	DBTx         pgx.Tx
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

func (e WithdrawalFromProcessingReceivedEvent) GetDatabaseTx() pgx.Tx {
	return e.DBTx
}

func (e WithdrawalFromProcessingReceivedEvent) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(&e); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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
