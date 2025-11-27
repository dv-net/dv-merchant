package aml_requests

import "github.com/dv-net/dv-merchant/internal/models"

type UserAMLKey struct {
	Name  models.AmlKeyType `json:"name" validate:"required,oneof=access_key_id access_key secret_key access_id"`
	Value *string           `json:"value,omitempty"`
} //	@name	UserAMLKey

type UpdateUserAMLKeys struct {
	Keys []UserAMLKey `json:"keys" required:"true" validate:"required,dive"`
} //	@name	UpdateUserAMLKeys
