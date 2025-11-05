// todo all event types create in this package
package types

import (
	"fmt"

	"github.com/dv-net/dv-merchant/internal/constant"
	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/jackc/pgx/v5"
)

const (
	ChangeInvoiceStausEventType = "invoice_update_status"
)

type InvoiceEvent interface {
	EventType() string
	GetStore() models.Store
	GetDatabaseTx() pgx.Tx
	GetWebhookEvent() models.WebhookEvent
	GetInvoice() models.Invoice
}

type ChangeInvoiceStatusEvent struct {
	Invoice      *models.Invoice
	Store        *models.Store
	Status       constant.InvoiceStatus
	Transactions []*models.Transaction
	WebhookEvent models.WebhookEvent
	DBTx         pgx.Tx
}

func (e ChangeInvoiceStatusEvent) Type() event.Type {
	return ChangeInvoiceStausEventType
}

func (e ChangeInvoiceStatusEvent) String() string {
	return fmt.Sprintf("ChangeInvoiceStatus: store=%d, webhook_event=%v", e.Store.ID, e.WebhookEvent)
}

func (e ChangeInvoiceStatusEvent) EventType() string {
	return ChangeInvoiceStausEventType
}

func (e ChangeInvoiceStatusEvent) GetStore() models.Store {
	return *e.Store
}

func (e ChangeInvoiceStatusEvent) GetDatabaseTx() pgx.Tx {
	return e.DBTx
}

func (e ChangeInvoiceStatusEvent) GetWebhookEvent() models.WebhookEvent {
	return e.WebhookEvent
}

func (e ChangeInvoiceStatusEvent) GetInvoice() models.Invoice {
	return *e.Invoice
}
