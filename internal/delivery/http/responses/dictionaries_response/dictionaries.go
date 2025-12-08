package dictionaries_response

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/settings_response"
	"github.com/dv-net/dv-merchant/internal/models"
)

type GetDictionariesResponse struct {
	BackendVersionTag     string                               `json:"backend_version_tag"`
	BackendVersionHash    string                               `json:"backend_version_hash"`
	ProcessingVersionTag  string                               `json:"processing_version_tag"`
	ProcessingVersionHash string                               `json:"processing_version_hash"`
	AvailableCurrencies   []*models.CurrencyShort              `json:"available_currencies"`
	AvailableRateSources  []string                             `json:"available_rate_sources"`
	AvailableAMLProviders []models.AMLSlug                     `json:"available_aml_providers"`
	BackendAddress        string                               `json:"backend_address" format:"ipv4"`
	GeneralSettings       []*settings_response.SettingResponse `json:"general_settings"`
} //	@name	GetDictionariesResponse
