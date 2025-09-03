package converters

import (
	"strings"

	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/notification_responses"
	stdmodels "github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/notification_settings"
	"github.com/dv-net/dv-merchant/internal/service/notify"
)

func FromNotificationList(dto []notification_settings.UserNotification) []notification_responses.UserNotificationResponse {
	res := make([]notification_responses.UserNotificationResponse, 0, len(dto))
	for _, notification := range dto {
		res = append(res, notification_responses.UserNotificationResponse{
			ID:           notification.ID,
			Name:         notification.Name,
			Category:     notification.Category,
			EmailEnabled: notification.EmailEnabled,
			TgEnabled:    notification.TgEnabled,
		})
	}

	return res
}

func FromNotificationHistoryList(models []*stdmodels.NotificationSendHistory) []notification_responses.NotificationHistoryResponse {
	res := make([]notification_responses.NotificationHistoryResponse, 0, len(models))
	for _, model := range models {
		response := notification_responses.NotificationHistoryResponse{
			ID:          model.ID,
			Destination: model.Destination,
			Sender:      model.Sender,
			CreatedAt:   model.CreatedAt.Time,
			UpdatedAt:   model.UpdatedAt.Time,
			Type:        model.Type,
			Channel:     model.Channel,
		}

		if model.MessageText.Valid {
			switch model.Channel {
			case stdmodels.EmailDeliveryChannel:
				msgText := extractHTML(model.MessageText.String)
				response.MessageText = &msgText
			case stdmodels.TelegramDeliveryChannel:
				response.MessageText = &model.MessageText.String
			}
		}

		if model.SentAt.Valid {
			response.SentAt = &model.SentAt.Time
		}

		res = append(res, response)
	}

	return res
}

func FromNotificationTypeList(types []notify.NotificationType) *notification_responses.NotificationTypeListResponse {
	res := &notification_responses.NotificationTypeListResponse{}

	for _, t := range types {
		res.Types = append(res.Types, struct {
			Label string `json:"label"`
			Value string `json:"value"`
		}{
			Label: t.Label,
			Value: t.Value,
		})
	}

	return res
}

func extractHTML(messageText string) string {
	// Find the first occurrence of '<' which indicates start of HTML
	index := strings.Index(messageText, "<")
	if index == -1 {
		return "" // No HTML found
	}
	return messageText[index:]
}
