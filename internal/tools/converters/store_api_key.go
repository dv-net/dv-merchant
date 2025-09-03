package converters

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/store_response"
	"github.com/dv-net/dv-merchant/internal/models"
)

func FromStoreAPIKeyModelToResponse(model *models.StoreApiKey) *store_response.StoreAPIKeyResponse {
	return &store_response.StoreAPIKeyResponse{
		ID:      model.ID.String(),
		Key:     model.Key,
		Enabled: model.Enabled,
	}
}

func FromStoreAPIKeyModelToResponses(models ...*models.StoreApiKey) []*store_response.StoreAPIKeyResponse {
	res := make([]*store_response.StoreAPIKeyResponse, 0, len(models))
	for _, model := range models {
		res = append(res, FromStoreAPIKeyModelToResponse(model))
	}
	return res
}
