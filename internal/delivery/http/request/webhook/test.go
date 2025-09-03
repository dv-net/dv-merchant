package webhook

import "github.com/google/uuid"

type TestWebhookRequest struct {
	WhID      uuid.UUID `json:"wh_id" validate:"required,uuid"`
	EventType string    `json:"event_type" validate:"required,oneof=PaymentReceived PaymentNotConfirmed WithdrawalFromProcessingReceived" enums:"PaymentReceived,PaymentNotConfirmed, WithdrawalFromProcessingReceived"`
} // @name TestWebhookRequest
