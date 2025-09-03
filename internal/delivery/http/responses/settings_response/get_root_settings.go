package settings_response

import (
	"time"
)

type GetRootSettingsResponse struct {
	Name      string    `json:"name"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
} // @name GetRootSettingsResponse
