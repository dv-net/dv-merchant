package responses

type GetNewVersionResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    GetNewVersionData `json:"data"`
}

type GetNewVersionData struct {
	Name             string `json:"name"`
	InstalledVersion string `json:"installed_version"`
	AvailableVersion string `json:"available_version"`
	NeedForUpdate    bool   `json:"need_for_update"`
}

type UpdateResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type GetUpdaterVersionResponse struct {
	AppVersion string `json:"app_version"`
	AppCommit  string `json:"app_commit"`
}
