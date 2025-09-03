package responses

import (
	htxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/models"
)

type (
	GetAPIKeyInformationResponse struct {
		Basic
		Data []*htxmodels.APIKeyInformation `json:"data,omitempty"`
	}
	GetUserUID struct {
		Basic
		Data int64 `json:"data,omitempty"`
	}
)
