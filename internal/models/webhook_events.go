package models

type WebhookEvent string //	@name	WebhookEvent

const (
	WebhookEventPaymentReceived                  WebhookEvent = "PaymentReceived"
	WebhookEventPaymentNotConfirmed              WebhookEvent = "PaymentNotConfirmed"
	WebhookEventWithdrawalFromProcessingReceived WebhookEvent = "WithdrawalFromProcessingReceived"
	WebhookEventPaymentAMLBlocked                WebhookEvent = "PaymentAMLBlocked"
)

func (s WebhookEvent) String() string {
	return string(s)
}
