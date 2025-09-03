package admin_requests

import "github.com/google/uuid"

type InitAuthRequest struct {
	OwnerID uuid.UUID `json:"owner_id"`
}
