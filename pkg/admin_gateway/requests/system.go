package admin_requests

import (
	"time"
)

type HeartBeatRequest struct {
	TS            time.Time      `json:"ts"`
	AnalyticsData *AnalyticsData `json:"analytics_data,omitempty"`
}
