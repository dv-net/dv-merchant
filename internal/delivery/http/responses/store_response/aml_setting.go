package store_response

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"
	"github.com/google/uuid"
)

type StoreAMLSettingsResponse struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	StoreID       uuid.UUID  `db:"store_id" json:"store_id"`
	Enabled       bool       `db:"enabled" json:"enabled"`
	RiskThreshold int32      `db:"risk_threshold" json:"risk_threshold"`
	ProviderSlug  string     `db:"provider_slug" json:"provider_slug"`
	CreatedAt     *time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     *time.Time `db:"updated_at" json:"updated_at"`
}

func NewStoreAMLSettingsResponse(amlSettings *models.StoreAmlSetting) StoreAMLSettingsResponse {
	return StoreAMLSettingsResponse{
		ID:            amlSettings.ID,
		StoreID:       amlSettings.StoreID,
		Enabled:       amlSettings.Enabled,
		RiskThreshold: amlSettings.RiskThreshold,
		ProviderSlug:  amlSettings.ProviderSlug.String(),
		CreatedAt:     pgtypeutils.DecodeTimeTz(amlSettings.CreatedAt),
		UpdatedAt:     pgtypeutils.DecodeTimeTz(amlSettings.UpdatedAt),
	}
}
