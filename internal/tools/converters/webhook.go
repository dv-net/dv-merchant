package converters

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/store_response"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/webhook_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/webhook"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_webhook_send_histories"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
)

func FromStoreWebhookModelToResponse(webhook *models.StoreWebhook) *store_response.StoreWebhookResponse {
	events := make([]string, 0, len(webhook.Events))
	for _, v := range webhook.Events {
		events = append(events, v.String())
	}

	return &store_response.StoreWebhookResponse{
		ID:      webhook.ID.String(),
		URL:     webhook.Url,
		Enabled: webhook.Enabled,
		Events:  events,
	}
}

func FromStoreWebhookModelToResponses(webhook ...*models.StoreWebhook) []*store_response.StoreWebhookResponse {
	webhooks := make([]*store_response.StoreWebhookResponse, 0, len(webhook))
	for _, v := range webhook {
		webhooks = append(webhooks, FromStoreWebhookModelToResponse(v))
	}
	return webhooks
}

func FromHistoryDataListToResponseList(data *storecmn.FindResponseWithPagingFlag[*repo_webhook_send_histories.FindRow]) webhook_response.WhHistoryResponse {
	return webhook_response.WhHistoryResponse{
		Items:          FromHistoryDtoListToResponseItemList(data.Items),
		NextPageExists: data.IsNextPageExists,
	}
}

func FromHistoryDtoListToResponseItemList(dtos []*repo_webhook_send_histories.FindRow) []webhook_response.WhHistory {
	res := make([]webhook_response.WhHistory, 0, len(dtos))
	for _, dto := range dtos {
		res = append(res, webhook_response.WhHistory{
			ID:            dto.ID,
			TransactionID: dto.TxID,
			StoreID:       dto.StoreID,
			URL:           dto.Url,
			CreatedAt:     dto.CreatedAt.Time,
			IsSuccess:     dto.ResponseStatusCode >= 200 && dto.ResponseStatusCode < 300,
			Request:       string(dto.Request),
			Response:      dto.Response,
			StatusCode:    dto.ResponseStatusCode,
		})
	}

	return res
}

func FromSendResultDtoToResponse(dto webhook.Result) webhook_response.SendTestResult {
	return webhook_response.SendTestResult{
		ResponseStatus: dto.Status,
		ResponseBody:   dto.Response,
		RequestBody:    dto.Request,
		ResponseCode:   dto.ResponseStatusCode,
	}
}
