package aml_requests

import "github.com/dv-net/dv-merchant/internal/models"

type GetHistoryRequest struct {
	ProviderSlug *models.AMLSlug `json:"provider_slug" validate:"omitempty,oneof=bit_ok aml_bot"`
	DateFrom     *string         `json:"date_from,omitempty"`
	DateTo       *string         `json:"date_to,omitempty"`
	Page         *uint32         `json:"page" validate:"omitempty,numeric,gte=1"`
	PageSize     *uint32         `json:"page_size" validate:"omitempty,min=1,max=100"`
} // @name GetHistoryRequest
