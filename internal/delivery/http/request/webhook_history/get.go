package webhook_history

import (
	"github.com/google/uuid"
)

type GetWhHistoryRequest struct {
	StoreUUIDs []uuid.UUID `json:"store_uuids" query:"store_uuids" validate:"omitempty"` //nolint:tagliatelle
	Page       *uint32     `json:"page" query:"page"`
	PageSize   *uint32     `json:"page_size" query:"page_size"`
} // @name GetWhHistoryRequest
