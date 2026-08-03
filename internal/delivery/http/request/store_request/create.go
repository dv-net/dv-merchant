package store_request

type CreateRequest struct {
	Site        *string `json:"site" validate:"omitempty,min=5,max=255"`
	Name        string  `json:"name" validate:"required,min=2,max=32"`
	Description *string `json:"description" validate:"omitempty,min=1,max=255"`
} //	@name	CreateStoreRequest
