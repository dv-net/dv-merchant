package store_webhook_request

import (
	"github.com/dv-net/dv-merchant/internal/models"
)

type UpdateRequest struct {
	URL     string                 `db:"url" json:"url" validate:"required,http_url"`
	Enabled bool                   `db:"enabled" json:"enabled"  validate:"boolean"`
	Events  []*models.WebhookEvent `db:"events" json:"events" validate:"required,dive"`
} // @name UpdateStoreWebhookRequest
