package admin_request

type RejectStoreRequest struct {
	Reason string `json:"reason" validate:"required"`
} //	@name	RejectStoreRequest

type ClarificationStoreRequest struct {
	Reason string `json:"reason" validate:"required"`
} //	@name	ClarificationStoreRequest
