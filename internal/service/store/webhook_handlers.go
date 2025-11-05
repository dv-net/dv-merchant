package store

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/event"
	eventtypes "github.com/dv-net/dv-merchant/internal/event/types"
	"github.com/dv-net/dv-merchant/internal/service/transactions"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_store_currencies"
	"github.com/google/uuid"
)

func (s *Service) handleDepositReceived(ev event.IEvent) error {
	convertedEv, ok := ev.(transactions.TransactionEvent)
	if !ok {
		return fmt.Errorf("invalid event type %s", ev.Type())
	}

	payload, err := s.prepareDepositHookPayload(
		convertedEv.GetTx(),
		convertedEv.GetCurrency(),
		convertedEv.GetWebhookEvent(),
		convertedEv.GetStoreExternalID(),
	)
	if err != nil {
		return fmt.Errorf("prepare deposit hook payload: %w", err)
	}

	if convertedEv.GetStore().MinimalPayment.GreaterThan(convertedEv.GetTx().GetAmountUsd()) {
		return nil
	}
	params := repo_store_currencies.FindByStoreIDParams{
		StoreID:    convertedEv.GetStore().ID,
		CurrencyID: convertedEv.GetCurrency().ID,
	}

	_, err = s.storage.StoreCurrencies().FindByStoreID(context.Background(), params)
	if err != nil {
		s.log.Errorw("store available currency not found", "error", err)
		return nil
	}

	return s.processWebhooksByEvent(ev, transactions.DepositReceivedEventType, payload)
}

func (s *Service) handleWithdrawalReceived(ev event.IEvent) error {
	convertedEv, ok := ev.(transactions.WithdrawalFromProcessingReceivedEvent)
	if !ok {
		return fmt.Errorf("invalid event type %s", ev.Type())
	}

	preparedPayload, err := s.prepareWithdrawalFromProcessingPayload(
		&convertedEv.Tx,
		&convertedEv.Currency,
		convertedEv.WithdrawalID,
		convertedEv.GetWebhookEvent(),
	)
	if err != nil {
		return fmt.Errorf("prepare withdrawal hook payload: %w", err)
	}

	return s.processWebhooksByEvent(ev, transactions.WithdrawalFromProcessingReceivedEventType, preparedPayload)
}

func (s *Service) handleInvoiceChangeStatus(ev event.IEvent) error {
	convertedEv, ok := ev.(eventtypes.ChangeInvoiceStatusEvent)
	if !ok {
		return fmt.Errorf("invalid event type %s", ev.Type())
	}

	preparedPayload, err := s.prepareInvoiceChangeStatusPayload(
		convertedEv.Invoice,
		convertedEv.Transactions,
		convertedEv.GetWebhookEvent(),
	)
	if err != nil {
		return fmt.Errorf("prepare invoice status changed webhook payload: %w", err)
	}

	return s.processWebhooksByInvoiceEvent(ev, preparedPayload)
}

func (s *Service) processWebhooksByEvent(
	ev event.IEvent,
	eventType string,
	hookPayload []byte,
) error {
	txCreatedEvent, ok := ev.(transactions.TransactionEvent)
	if !ok || txCreatedEvent.EventType() != eventType {
		return nil
	}

	params := webhookProcessingParams{
		StoreID:      txCreatedEvent.GetStore().ID,
		EntityID:     uuid.NullUUID{UUID: txCreatedEvent.GetTx().GetID(), Valid: true},
		WebhookType:  string(txCreatedEvent.GetWebhookEvent()),
		WebhookEvent: string(txCreatedEvent.GetWebhookEvent()),
		DBTx:         txCreatedEvent.GetDatabaseTx(),
		Payload:      hookPayload,
		LogContext:   "transaction",
		LogFields: map[string]interface{}{
			"tx_hash": txCreatedEvent.GetTx().GetTxHash(),
		},
	}

	return s.processWebhooksCommon(params)
}

func (s *Service) processWebhooksByInvoiceEvent(
	ev event.IEvent,
	hookPayload []byte,
) error {
	invoiceEvent, ok := ev.(eventtypes.ChangeInvoiceStatusEvent)
	fmt.Println("invoiceEvent %w", invoiceEvent)
	if !ok {
		return fmt.Errorf("invalid event type %s, expected ChangeInvoiceStatusEvent", ev.Type())
	}

	webhookType := fmt.Sprintf("%s:%s",
		invoiceEvent.GetWebhookEvent(),
		invoiceEvent.Status,
	)

	params := webhookProcessingParams{
		StoreID:      invoiceEvent.Store.ID,
		EntityID:     uuid.NullUUID{UUID: invoiceEvent.Invoice.ID, Valid: true},
		WebhookType:  webhookType,
		WebhookEvent: string(invoiceEvent.GetWebhookEvent()),
		DBTx:         invoiceEvent.DBTx,
		Payload:      hookPayload,
		LogContext:   "invoice",
		LogFields: map[string]interface{}{
			"invoice_id":     invoiceEvent.Invoice.ID.String(),
			"invoice_status": string(invoiceEvent.Status),
			"order_id":       invoiceEvent.Invoice.OrderID,
		},
	}

	return s.processWebhooksCommon(params)
}
