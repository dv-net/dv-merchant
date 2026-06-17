package store

import "github.com/dv-net/dv-merchant/internal/models"

type UpdateStoreCurrencyDTO struct {
	CurrencyIDs []string `json:"currency_ids"` //nolint:tagliatelle
}

type UpdateAMLSettingsDTO struct {
	Enabled       bool  `json:"enabled"`
	RiskThreshold int32 `json:"risk_threshold"`
	ProviderSlug  *models.AMLSlug
}
