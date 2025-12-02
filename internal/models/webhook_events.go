package models

type WebhookEvent string //	@name	WebhookEvent

const (
	WebhookEventPaymentReceived                  WebhookEvent = "PaymentReceived"
	WebhookEventPaymentNotConfirmed              WebhookEvent = "PaymentNotConfirmed"
	WebhookEventWithdrawalFromProcessingReceived WebhookEvent = "WithdrawalFromProcessingReceived"
)

func (s WebhookEvent) String() string {
	return string(s)
}
