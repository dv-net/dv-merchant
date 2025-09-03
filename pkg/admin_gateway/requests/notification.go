package admin_requests

type NotificationIdentity struct {
	Destination string `json:"destination"`
	Channel     string `json:"channel,omitempty"`
}

type VerifyNotification struct {
	BackendClientID string               `json:"_backend_client_id"` //nolint:tagliatelle
	BackendDomain   string               `json:"_backend_domain"`    //nolint:tagliatelle
	Locale          string               `json:"locale"`
	Identity        NotificationIdentity `json:"identity"`
	Code            string               `json:"code"`
}

type WalletsRequestNotification struct {
	BackendClientID string               `json:"_backend_client_id"` //nolint:tagliatelle
	BackendDomain   string               `json:"_backend_domain"`    //nolint:tagliatelle
	Locale          string               `json:"locale"`
	Identity        NotificationIdentity `json:"identity"`
	Addresses       []WalletAddressDTO   `json:"addresses"`
}

type WalletAddressDTO struct {
	CurrencyID string `json:"currency_id"`
	Blockchain string `json:"blockchain"`
	Address    string `json:"address"`
}
