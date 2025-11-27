package statistics_request

import (
	"github.com/dv-net/dv-merchant/internal/service/transactions"

	"github.com/google/uuid"
)

type GetDepositsSummaryRequest struct {
	DateFrom    *string                            `json:"date_from" query:"date_from" format:"date-time"`
	DateTo      *string                            `json:"date_to" query:"date_to" format:"date-time"`
	Resolution  *transactions.StatisticsResolution `json:"resolution" query:"resolution" validate:"omitempty,oneof=hour day week month quarter year"`
	CurrencyIDs []string                           `json:"currency_ids" query:"currency_ids"` //nolint:tagliatelle
	StoreUUIDs  []uuid.UUID                        `json:"store_uuids" query:"store_uuids"`   //nolint:tagliatelle
} //	@name	GetDepositsSummaryRequest
