package user_request

type ConfirmTgUnlinkRequest struct {
	Code string `json:"code" validate:"required"`
} //	@name	ConfirmTgUnlinkRequest
