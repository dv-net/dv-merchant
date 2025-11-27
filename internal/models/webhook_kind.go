package models

type WebhookKind string //	@name	WebhookKind

const (
	WebhookKindTransfer       WebhookKind = "transfer"
	WebhookKindDeposit        WebhookKind = "deposit"
	WebhookKindTransferStatus WebhookKind = "transfer_status"
)
