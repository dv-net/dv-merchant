package updater

type ApplicationVersion struct {
	BackendVersion    *VersionInfo `json:"backend_version"`
	ProcessingVersion *VersionInfo `json:"processing_version"`
}

type VersionInfo struct {
	Name             string `json:"name"`
	AvailableVersion string `json:"available_version"`
	InstalledVersion string `json:"installed_version"`
	NeedForUpdate    bool   `json:"need_for_update"`
}
