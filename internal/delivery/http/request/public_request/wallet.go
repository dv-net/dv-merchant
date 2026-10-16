package public_request

type GetAmlCheckRequest struct {
	Hash string `json:"hash" query:"hash" validate:"required"`
} //	@name	GetAmlCheckRequest
