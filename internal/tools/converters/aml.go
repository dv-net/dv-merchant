package converters

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/aml_requests"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/aml_responses"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/aml"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_checks"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
)

func ConvertAmlKeysToResponseKeys(keysDTO *aml.UserKeysDTO) []aml_responses.AMLKey {
	responseKeys := make([]aml_responses.AMLKey, 0, len(keysDTO.Keys))
	for _, key := range keysDTO.Keys {
		responseKeys = append(responseKeys, aml_responses.AMLKey{
			Name:  string(key.Name),
			Value: key.Value,
		})
	}

	return responseKeys
}

func ConvertAMLKeysRequestToDTO(slug models.AMLSlug, req *aml_requests.UpdateUserAMLKeys) aml.UserKeysDTO {
	userKeysDTO := aml.UserKeysDTO{
		Slug: slug,
		Keys: make([]aml.UserKeyDTO, 0, len(req.Keys)),
	}
	for _, key := range req.Keys {
		userKeysDTO.Keys = append(userKeysDTO.Keys, aml.UserKeyDTO{
			Name:  key.Name,
			Value: key.Value,
		})
	}

	return userKeysDTO
}

func GetAMLCheckHistoryResponse(m *storecmn.FindResponseWithFullPagination[*repo_aml_checks.FindRow]) *storecmn.FindResponseWithFullPagination[*aml_responses.AmlHistoryResponse] {
	items := make([]*aml_responses.AmlHistoryResponse, 0, len(m.Items))
	for _, v := range m.Items {
		item := &aml_responses.AmlHistoryResponse{
			ID:          v.ID,
			UserID:      v.UserID,
			ServiceID:   v.ServiceID,
			ServiceSlug: v.Slug,
			ExternalID:  v.ExternalID,
			Status:      v.Status,
			Score:       v.Score,
			RiskLevel:   v.RiskLevel,
		}

		if v.CreatedAt.Valid {
			createdAt := v.CreatedAt.Time
			item.CreatedAt = &createdAt
		}
		if v.UpdatedAt.Valid {
			updatedAt := v.UpdatedAt.Time
			item.UpdatedAt = &updatedAt
		}

		requestHistory := make([]aml_responses.CheckHistory, 0, len(v.History))
		for _, h := range v.History {
			historyItem := aml_responses.CheckHistory{
				ID:              h.ID,
				AmlCheckID:      h.AmlCheckID,
				RequestPayload:  string(h.RequestPayload),
				ServiceResponse: string(h.ServiceResponse),
				AttemptNumber:   h.AttemptNumber,
			}

			if h.ErrorMsg.Valid {
				errorMsg := h.ErrorMsg.String
				historyItem.ErrorMsg = &errorMsg
			}

			if h.CreatedAt.Valid {
				createdAt := h.CreatedAt.Time
				historyItem.CreatedAt = &createdAt
			}
			if h.UpdatedAt.Valid {
				updatedAt := h.UpdatedAt.Time
				historyItem.UpdatedAt = &updatedAt
			}

			requestHistory = append(requestHistory, historyItem)
		}

		item.RequestHistory = requestHistory
		items = append(items, item)
	}

	return &storecmn.FindResponseWithFullPagination[*aml_responses.AmlHistoryResponse]{
		Items:      items,
		Pagination: m.Pagination,
	}
}
