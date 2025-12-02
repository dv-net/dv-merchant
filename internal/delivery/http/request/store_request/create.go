package store_request

type CreateRequest struct {
	Site *string `json:"site" validate:"omitempty,min=5,max=255"`
	Name string  `db:"name" json:"name" validate:"required,min=2,max=32"`
} //	@name	CreateStoreRequest
