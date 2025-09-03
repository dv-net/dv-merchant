package notification_request

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
)

type GetNotificationHistoryRequest struct {
	Page         *uint32                   `json:"page" query:"page"`
	PageSize     *uint32                   `json:"page_size" query:"page_size"`
	IDs          []uuid.UUID               `json:"ids" query:"ids" validate:"omitempty,uuid"` //nolint:tagliatelle
	Destinations []string                  `json:"destinations" query:"destinations" validate:"omitempty,unique"`
	Types        []models.NotificationType `json:"types" query:"types" validate:"omitempty,unique"`
	Channels     []models.DeliveryChannel  `json:"channels" query:"channels" validate:"omitempty,unique"`
	CreatedFrom  *time.Time                `json:"created_from" query:"created_from" validate:"omitempty"`
	CreatedTo    *time.Time                `json:"created_to" query:"created_to" validate:"omitempty"`
	SentFrom     *time.Time                `json:"sent_from" query:"sent_from" validate:"omitempty"`
	SentTo       *time.Time                `json:"sent_to" query:"sent_to" validate:"omitempty"`
} // @name GetNotificationHistoryRequest
