package admin_request

type GetUsersRequest struct {
	Roles    *string `json:"roles" query:"roles" enums:"root,user"`
	Page     *uint32 `json:"page" query:"page"`
	PageSize *uint32 `json:"page_size" query:"page_size"`
} // @name GetUsersRequest
