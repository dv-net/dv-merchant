package store_response

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"
	"github.com/google/uuid"
)

type StoreAMLSettingsResponse struct {
	ID                      uuid.UUID  `json:"id"`
	StoreID                 uuid.UUID  `json:"store_id"`
	Enabled                 bool       `json:"enabled"`
	RiskThreshold           int32      `json:"risk_threshold"`
	ProviderSlug            string     `json:"provider_slug"`
	IgnoredSignalCategories []string   `json:"ignored_signal_categories"`
	CreatedAt               *time.Time `json:"created_at"`
	UpdatedAt               *time.Time `json:"updated_at"`
}

func NewStoreAMLSettingsResponse(amlSettings *models.StoreAmlSetting) StoreAMLSettingsResponse {
	categories, err := amlSettings.ParseIgnoredSignalCategories()
	if err != nil {
		categories = nil
	}
	return StoreAMLSettingsResponse{
		ID:                      amlSettings.ID,
		StoreID:                 amlSettings.StoreID,
		Enabled:                 amlSettings.Enabled,
		RiskThreshold:           amlSettings.RiskThreshold,
		ProviderSlug:            amlSettings.ProviderSlug.String(),
		IgnoredSignalCategories: categories,
		CreatedAt:               pgtypeutils.DecodeTimeTz(amlSettings.CreatedAt),
		UpdatedAt:               pgtypeutils.DecodeTimeTz(amlSettings.UpdatedAt),
	}
}
