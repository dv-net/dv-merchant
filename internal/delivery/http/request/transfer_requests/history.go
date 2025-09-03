package transfer_requests

type TransferHistoryRequest struct {
	Address  string  `json:"address" query:"address"`
	Page     *uint32 `json:"page" query:"page"`
	PageSize *uint32 `json:"page_size" query:"page_size"`
} // @name TransferHistoryRequest
