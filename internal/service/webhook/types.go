package webhook

import (
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_webhook_send_queue"

	"github.com/google/uuid"
)

type StoreWhResponse struct {
	Success bool `json:"success"`
}

type PreparedHookDto struct {
	ID            uuid.NullUUID
	TransactionID uuid.UUID
	StoreID       uuid.UUID
	IsManual      bool
	Event         string
	Payload       []byte
	Signature     string
	URL           string
	RetriesCount  int64
}

type Message struct {
	TxID      uuid.UUID `json:"tx_id"`
	WebhookID uuid.UUID `json:"webhook_id"`
	Type      string    `json:"type"`
	Data      []byte    `json:"data"`
	Delay     int16     `json:"delay"`
	Signature string    `json:"signature"`
}

type Result struct {
	Status             string
	Response           string
	Request            string
	ResponseStatusCode int
}

func prepareHookDtoByRaw(v *repo_webhook_send_queue.GetQueuedWebhooksRow) PreparedHookDto {
	return PreparedHookDto{
		ID: uuid.NullUUID{
			UUID:  v.ID,
			Valid: true,
		},
		TransactionID: v.TransactionID,
		StoreID:       v.StoreID,
		Event:         v.Event,
		Payload:       v.Payload,
		Signature:     v.Signature,
		URL:           v.Url,
		RetriesCount:  v.RetriesCount,
	}
}
