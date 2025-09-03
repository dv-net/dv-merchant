package log_response

import (
	"time"

	"github.com/google/uuid"
)

type LogTypeData struct {
	ID        int32     `json:"id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at" format:"date-time"`
} // @name MonitorTypeData

type GetLogTypesResponse struct {
	Items []LogTypeData `json:"items"`
} // @name GetMonitorTypesResponse

type LogData struct {
	ProcessID uuid.UUID                `json:"process_id"`
	Failure   bool                     `json:"failure"`
	CreatedAt time.Time                `json:"created_at"`
	Messages  []map[string]interface{} `json:"messages"`
} // name LogData

type GetLogsResponse struct {
	Items []LogData `json:"items"`
} // @name GetMonitorTypesResponse
