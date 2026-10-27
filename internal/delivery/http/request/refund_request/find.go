package refund_request

import (
	"github.com/dv-net/dv-merchant/internal/models"
)

type FindRequest struct {
	Statuses           []string           `json:"statuses" query:"statuses" validate:"omitempty,dive,oneof=pending_review rejected processing scheduled canceled completed failed"`
	Blockchain         *models.Blockchain `json:"blockchain" query:"blockchain"`
	TxHash             string             `json:"tx_hash" query:"tx_hash"` // deposit or refund transfer hash
	FromAddress        string             `json:"from_address" query:"from_address"`
	DestinationAddress string             `json:"destination_address" query:"destination_address"`
	DateFrom           *string            `json:"date_from" query:"date_from" format:"date-time"`
	DateTo             *string            `json:"date_to" query:"date_to" format:"date-time"`
	Page               *uint32            `json:"page" query:"page" validate:"omitempty,min=1"`
	PageSize           *uint32            `json:"page_size" query:"page_size" validate:"omitempty,min=1,max=100"`
	SortBy             *string            `json:"sort_by" query:"sort_by" validate:"omitempty,oneof=created_at updated_at score amount_usd" enums:"created_at,updated_at,score,amount_usd"`
	SortDirection      *string            `json:"sort_direction" query:"sort_direction" validate:"omitempty,oneof=asc desc" enums:"asc,desc"`
} //	@name	FindRefundRequestsRequest
