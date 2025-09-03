package store_api_key_request

type UpdateStatusRequest struct {
	Status bool `db:"status" json:"status" validate:"boolean"`
} // @name UpdateAPIKeyStatusRequest
