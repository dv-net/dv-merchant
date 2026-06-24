package store_aml_request

import "github.com/dv-net/dv-merchant/internal/models"

type UpdateStoreAMLSettingsRequest struct {
	Enabled       bool            `json:"enabled"`
	RiskThreshold int32           `json:"risk_threshold" validate:"min=0,max=100"`
	ProviderSlug  *models.AMLSlug `json:"provider_slug"`
} //	@name	UpdateStoreAMLSettingsRequest
