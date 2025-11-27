package webhook_response

import (
	"time"

	"github.com/google/uuid"
)

type WhHistoryResponse struct {
	Items          []WhHistory `json:"items"`
	NextPageExists bool        `json:"next_page_exists"`
} //	@name	WhHistoryResponse

type WhHistory struct {
	ID            uuid.UUID     `json:"id"`
	StoreID       uuid.NullUUID `json:"store_id"`
	TransactionID uuid.UUID     `json:"transaction_id"`
	CreatedAt     time.Time     `json:"created_at" format:"date-time"`
	IsSuccess     bool          `json:"is_success"`
	Request       string        `json:"request"`
	Response      *string       `json:"response"`
	StatusCode    int           `json:"status_code"`
	URL           string        `json:"url"`
} //	@name	WhHistory
