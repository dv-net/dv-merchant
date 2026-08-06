package store

import (
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/google/uuid"
)

type UpdateStoreCurrencyDTO struct {
	CurrencyIDs []string `json:"currency_ids"` //nolint:tagliatelle
}

type UpdateAMLSettingsDTO struct {
	Enabled                 bool  `json:"enabled"`
	RiskThreshold           int32 `json:"risk_threshold"`
	ProviderSlug            *models.AMLSlug
	IgnoredSignalCategories []string `json:"ignored_signal_categories"`
}

type VerifyStoreDTO struct {
	StoreID uuid.UUID    `json:"store_id"`
	Admin   *models.User `json:"admin"`
}

type RejectStoreDTO struct {
	StoreID uuid.UUID    `json:"store_id"`
	Admin   *models.User `json:"admin"`
	Reason  string       `json:"reason"`
}

type ResendStoreVerificationDTO struct {
	StoreID uuid.UUID `json:"store_id"`
	Comment string    `json:"comment"`
}
type ClarificationStoreDTO struct {
	StoreID uuid.UUID    `json:"store_id"`
	Admin   *models.User `json:"admin"`
	Reason  string       `json:"reason"`
}
