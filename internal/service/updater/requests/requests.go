package requests

type UpdateRequest struct {
	Name string `json:"name"  validate:"required"`
}
