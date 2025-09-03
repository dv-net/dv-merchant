package notification_responses

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
)

type NotificationHistoryResponse struct {
	ID          uuid.UUID               `json:"id"`
	Destination string                  `json:"destination"`
	MessageText *string                 `json:"message_text"`
	Sender      string                  `json:"sender"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
	SentAt      *time.Time              `json:"sent_at"`
	Type        models.NotificationType `json:"type"`
	Channel     models.DeliveryChannel  `json:"channel"`
} // @name NotificationHistoryResponse

type NotificationTypeListResponse struct {
	Types []struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"types"`
} // @name NotificationTypeListResponse
