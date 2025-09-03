package transactions_request

import (
	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/shopspring/decimal"
)

type GetStatistics struct {
	Currencies   []string                 `json:"currencies" query:"currencies"`
	Resolution   string                   `json:"resolution" query:"resolution" validate:"required,oneof=day hour"`
	Type         *models.TransactionsType `json:"type" query:"type" validate:"omitempty,oneof=deposit transfer" enums:"deposit,transfer"`
	IsSystem     bool                     `json:"is_system" query:"is_system"`
	MinAmountUSD decimal.Decimal          `json:"min_amount_usd" query:"min_amount_usd"`
	Blockchain   *models.Blockchain       `json:"blockchain" query:"blockchain"`
	DateFrom     *string                  `json:"date_from" query:"date_from" format:"date-time"`
	DateTo       *string                  `json:"date_to" query:"date_to" format:"date-time"`
} // @name GetStatistics
