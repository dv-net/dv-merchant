package system_response

import "github.com/dv-net/dv-merchant/internal/service/updater"

type VersionResponse struct {
	NewBackendVersion    *updater.VersionInfo `json:"new_backend_version"`
	NewProcessingVersion *updater.VersionInfo `json:"new_processing_version"`
} // @name VersionResponse
