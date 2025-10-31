package models

type WebhookEvent string // @name WebhookEvent

const (
	WebhookEventPaymentReceived                  WebhookEvent = "PaymentReceived"
	WebhookEventPaymentNotConfirmed              WebhookEvent = "PaymentNotConfirmed"
	WebhookEventWithdrawalFromProcessingReceived WebhookEvent = "WithdrawalFromProcessingReceived"
	WebhookEventInvoiceChangeStatus              WebhookEvent = "InvoiceChangeStatus"
)

func (s WebhookEvent) String() string {
	return string(s)
}
