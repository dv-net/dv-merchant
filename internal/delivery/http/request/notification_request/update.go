package notification_request

import "github.com/google/uuid"

type Update struct {
	TgEnabled    bool `json:"tg_enabled"`
	EmailEnabled bool `json:"email_enabled"`
} // @name Update

type UpdateList struct {
	List []struct {
		ID           uuid.UUID `json:"id" validate:"required"`
		TgEnabled    bool      `json:"tg_enabled"`
		EmailEnabled bool      `json:"email_enabled"`
	} `json:"list" validate:"dive,required"`
} // @name UpdateList

type TestNotificationRequest struct {
	Recipient string `json:"recipient" validate:"required"`
} // @name TestNotificationRequest
