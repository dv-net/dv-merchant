package admin_request

type GetStoresRequest struct {
	Status   *string `json:"status" query:"status" enums:"PENDING,SUCCESS,REJECTED"`
	Page     *uint32 `json:"page" query:"page"`
	PageSize *uint32 `json:"page_size" query:"page_size"`
} //	@name	GetStoresRequest
