package aml_requests

import "github.com/dv-net/dv-merchant/internal/models"

type ScoreTxRequest struct {
	TxID          string         `json:"tx_id" validate:"required"`
	CurrencyID    string         `json:"currency_id" validate:"required"`
	ProviderSlug  models.AMLSlug `json:"provider_slug" validate:"required"`
	Direction     string         `json:"direction" validate:"required,oneof=in out"`
	OutputAddress string         `json:"output_address" validate:"required"`
} //	@name	ScoreTxRequest
