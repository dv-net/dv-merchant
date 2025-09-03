package notification_settings

import "github.com/google/uuid"

type UserNotification struct {
	ID           uuid.UUID
	Name         string
	Category     string
	EmailEnabled bool
	TgEnabled    bool
}

type UpdateDTO struct {
	ID           uuid.UUID
	EmailEnabled bool
	TgEnabled    bool
}

func (ud UpdateDTO) IsFullDisable() bool {
	return !ud.EmailEnabled && !ud.TgEnabled
}
