package store_webhook_request

import (
	"github.com/dv-net/dv-merchant/internal/models"
)

type CreateRequest struct {
	URL     string                 `db:"url" json:"url" validate:"required,http_url" format:"url"`
	Enabled bool                   `db:"enabled" json:"enabled"`
	Events  []*models.WebhookEvent `db:"events" json:"events,omitempty" validate:"required,dive,oneof=PaymentReceived PaymentNotConfirmed WithdrawalFromProcessingReceived"`
} // @name CreateStoreWebhookRequest
