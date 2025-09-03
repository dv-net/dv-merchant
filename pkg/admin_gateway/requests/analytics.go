package admin_requests

import (
	"time"

	"github.com/shopspring/decimal"
)

type SystemInfo struct {
	OS             string    `json:"os"`
	Architecture   string    `json:"architecture"`
	Version        string    `json:"version"`
	GoVersion      string    `json:"go_version"`
	CollectionTime time.Time `json:"collection_time"`
}

type EnvironmentInfo struct {
	BackendVersionHash    string `json:"backend_version_hash"`
	BackendVersionTag     string `json:"backend_version_tag"`
	ProcessingVersionHash string `json:"processing_version_hash"`
	ProcessingVersionTag  string `json:"processing_version_tag"`
	UpdaterVersionHash    string `json:"updater_version_hash"`
	UpdaterVersionTag     string `json:"updater_version_tag"`
}

type TransferStatsByCurrency struct {
	CurrencyID     string          `json:"currency_id"`
	CurrencyCode   string          `json:"currency_code"`
	TotalCount     int64           `json:"total_count"`
	TotalAmount    decimal.Decimal `json:"total_amount"`
	TotalAmountUSD decimal.Decimal `json:"total_amount_usd"`
}

type AnalyticsData struct {
	SystemInfo      *SystemInfo               `json:"system_info,omitempty"`
	EnvironmentInfo *EnvironmentInfo          `json:"environment_info,omitempty"`
	TransferStats   []TransferStatsByCurrency `json:"transfer_stats,omitempty"`
}
