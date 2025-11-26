package notification_responses

import "github.com/google/uuid"

type UserNotificationResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	EmailEnabled bool      `json:"email_enabled"`
	TgEnabled    bool      `json:"tg_enabled"`
} //	@name	UserNotificationResponse
