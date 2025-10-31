package store

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/service/webhook"
	"github.com/dv-net/dv-merchant/internal/tools/hash"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type webhookProcessingParams struct {
	StoreID      uuid.UUID
	EntityID     uuid.NullUUID
	WebhookType  string
	WebhookEvent string
	DBTx         pgx.Tx
	Payload      []byte
	LogContext   string
	LogFields    map[string]interface{}
}

// processWebhooksCommon is a generalized method for processing webhooks
func (s *Service) processWebhooksCommon(params webhookProcessingParams) error {
	webhooks, err := s.getWebhooksByStore(
		context.Background(),
		params.StoreID,
		params.WebhookEvent,
		params.DBTx,
	)
	if err != nil {
		s.log.Errorw("store webhook not found", "error", err)
		return nil
	}
	for _, v := range webhooks {
		// Check if webhook was already sent
		if params.EntityID.Valid {
			if s.isWebhookAlreadySent(v.StoreWebhook.Url, params.WebhookType, params.EntityID.UUID, params.DBTx) {
				continue
			}
		}

		message := webhook.Message{
			TxID:      params.EntityID.UUID,
			WebhookID: v.StoreWebhook.ID,
			Type:      params.WebhookType,
			Data:      params.Payload,
			Signature: hash.SHA256Signature(params.Payload, v.Secret.String),
		}

		whSendErr := s.webhookService.Send(&message, params.DBTx)
		if whSendErr != nil {
			logFields := []interface{}{
				"error", whSendErr,
				"store_id", params.StoreID.String(),
				"wh_type", params.WebhookType,
				"wh_body", string(params.Payload),
			}

			for k, v := range params.LogFields {
				logFields = append(logFields, k, v)
			}

			s.log.Errorw(
				"store webhook send error ("+params.LogContext+")",
				logFields...,
			)
		}

	}

	return nil
}
