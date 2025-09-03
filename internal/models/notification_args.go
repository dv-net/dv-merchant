package models

import "github.com/google/uuid"

type NotificationArgs struct {
	UserID  *uuid.UUID `json:"user_id,omitempty"`
	StoreID *uuid.UUID `json:"store_id,omitempty"`
}
