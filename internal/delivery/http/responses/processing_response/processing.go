package processing_response

type InitProcessingResponse struct {
	BaseURL             string `json:"base_url" format:"uri"`
	ProcessingClientID  string `json:"processing_client_id" format:"uuid"`
	ProcessingClientKey string `json:"processing_client_key"`
} // @name InitProcessingResponse

type ProcessingListResponse struct{} // @name ProcessingListResponse

type ProcessingWalletResponse struct{} // @name ProcessingWalletResponse

type CallbackURLResponse struct {
	CallbackURL string `json:"callback_url"`
} // @name CallbackURLResponse

type OwnerProcessingResponse struct {
	OwnerID string `json:"owner_id"`
} // @name OwnerProcessingResponse
