package webhook

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/jackc/pgx/v5"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_stores"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_webhook_send_histories"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_webhook_send_queue"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/google/uuid"
)

const (
	WebhookSendStatusFailed  string = "failed"
	WebhookSendStatusSuccess string = "success"
	webhookPollingInterval          = time.Second * 2
)

type IWebHook interface {
	Run(ctx context.Context)
	Send(message *Message, dbTx pgx.Tx) error
	ProcessPlainMessage(ctx context.Context, dto PreparedHookDto) error
	SendWebhook(ctx context.Context, url string, payload []byte, sign string) (Result, error)
	GetHistory(ctx context.Context, user models.User, storeUUIDs []uuid.UUID, page, pageSize *uint32) (*storecmn.FindResponseWithPagingFlag[*repo_webhook_send_histories.FindRow], error)
}

type service struct {
	storage  storage.IStorage
	log      logger.Logger
	maxTries int
	locker   queueLocker
}

var _ IWebHook = (*service)(nil)

func New(c config.WebHook, s storage.IStorage, l logger.Logger) IWebHook {
	srv := service{
		storage:  s,
		log:      l,
		maxTries: c.MaxTries,
		locker: queueLocker{
			mu:           &sync.Mutex{},
			whInProgress: make(map[uuid.UUID]struct{}),
		},
	}

	return &srv
}

func (s *service) Send(message *Message, dbTx pgx.Tx) error {
	_, err := s.storage.WebHookSendQueue(repos.WithTx(dbTx)).Create(
		context.Background(),
		repo_webhook_send_queue.CreateParams{
			WebhookID:     message.WebhookID,
			SecondsDelay:  message.Delay,
			TransactionID: message.TxID,
			Payload:       message.Data,
			Signature:     message.Signature,
			Event:         message.Type,
		},
	)
	if err != nil {
		s.log.Errorw("create webhook_send_queue", "error", err)
		return fmt.Errorf("create webhook_send_queue: %w", err)
	}

	return nil
}

func (s *service) Run(ctx context.Context) {
	ticker := time.NewTicker(webhookPollingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			go func() {
				if err := s.processWhQueue(ctx); err != nil {
					s.log.Errorw("Precess webhook queue error", "error", err)
				}
			}()
		case <-ctx.Done():
			return
		}
	}
}

func (s *service) SendWebhook(ctx context.Context, url string, payload []byte, sign string) (Result, error) {
	result := Result{
		Status: WebhookSendStatusFailed,
	}
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, url, bytes.NewBuffer(payload),
	)
	if err != nil {
		return result, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Sign", sign)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.log.Errorw("webhook sent http", "error", err)
		return result, nil
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			s.log.Errorw("close wh resp body error", "error", err)
		}
	}()

	result.ResponseStatusCode = resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.log.Errorw("read web hook response body failed", "error", err)
	}

	whResp := &StoreWhResponse{}
	if err = json.NewDecoder(bytes.NewBuffer(body)).Decode(whResp); err != nil {
		s.log.Errorw("decode body failed", "error", err)
	}

	whStatus := WebhookSendStatusFailed
	if whResp.Success {
		whStatus = WebhookSendStatusSuccess
	}

	result.Status = whStatus
	result.Response = string(body)
	result.Request = string(payload)

	return result, nil
}

func (s *service) processWhQueue(ctx context.Context) error {
	webhooksData, err := s.storage.WebHookSendQueue().GetQueuedWebhooks(ctx)
	if err != nil {
		return fmt.Errorf("fetch wh queue: %w", err)
	}

	for _, v := range webhooksData {
		if v.LastSentAt.Valid && time.Now().Before(v.LastSentAt.Time.Add(time.Second*time.Duration(v.SecondsDelay))) {
			continue
		}

		if !s.locker.Acquire(v.WebhookID) {
			// Webhook currently processing from another goroutine
			continue
		}

		if err := s.ProcessPlainMessage(ctx, prepareHookDtoByRaw(v)); err != nil {
			s.log.Errorw("Processing wh message", "error", err)
		}

		s.locker.Release(v.WebhookID)
	}

	return nil
}

func (s *service) ProcessPlainMessage(ctx context.Context, dto PreparedHookDto) error {
	result, sendWhErr := s.SendWebhook(ctx, dto.URL, dto.Payload, dto.Signature)
	if sendWhErr != nil {
		s.log.Errorw("send webhook error", "error", sendWhErr)
	}

	if createHistoryErr := s.createSendHistory(ctx, dto, result); createHistoryErr != nil {
		s.log.Errorw("create send history failed", "error", createHistoryErr)
	}

	if dto.ID.Valid && !dto.IsManual {
		if result.Status == WebhookSendStatusSuccess || dto.RetriesCount >= int64(s.maxTries) {
			return s.storage.WebHookSendQueue().Delete(ctx, dto.ID.UUID)
		}

		return s.storage.WebHookSendQueue().UpdateDelay(ctx, repo_webhook_send_queue.UpdateDelayParams{
			ID:    dto.ID.UUID,
			Delay: int16(60 * math.Pow(float64(2), float64(dto.RetriesCount))),
		})
	}

	if sendWhErr != nil {
		return fmt.Errorf("send webhook failed: %w", sendWhErr)
	}

	return nil
}

func (s *service) GetHistory(
	ctx context.Context,
	user models.User,
	storeUUIDs []uuid.UUID,
	page, pageSize *uint32,
) (*storecmn.FindResponseWithPagingFlag[*repo_webhook_send_histories.FindRow], error) {
	storeIDs, err := s.storage.UserStores().GetStoreIDsByUser(ctx, repo_user_stores.GetStoreIDsByUserParams{
		UserID:     user.ID,
		StoreUuids: storeUUIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("fetch target stores: %w", err)
	}
	if len(storeIDs) == 0 {
		return &storecmn.FindResponseWithPagingFlag[*repo_webhook_send_histories.FindRow]{}, nil
	}

	commonParams := storecmn.NewCommonFindParams()
	commonParams.SetPage(page)
	commonParams.SetPageSize(pageSize)
	commonParams.SetIsAscOrdering(false)

	params := repo_webhook_send_histories.GetHistoriesParams{
		CommonFindParams: *commonParams,
		StoreIDs:         storeIDs,
	}
	params.Page = page
	params.PageSize = pageSize

	return s.storage.WebHookSendHistories().GetByStores(ctx, params)
}

func (s *service) createSendHistory(ctx context.Context, hookData PreparedHookDto, result Result) error {
	params := repo_webhook_send_histories.CreateParams{
		TxID:               hookData.TransactionID,
		SendQueueJobID:     hookData.ID,
		Type:               hookData.Event,
		Url:                hookData.URL,
		Status:             result.Status,
		Request:            hookData.Payload,
		Response:           &result.Response,
		ResponseStatusCode: result.ResponseStatusCode,
		StoreID:            uuid.NullUUID{UUID: hookData.StoreID, Valid: true},
		IsManual:           pgtype.Bool{Bool: hookData.IsManual, Valid: true},
	}
	if _, err := s.storage.WebHookSendHistories().Create(ctx, params); err != nil {
		return fmt.Errorf("save web hook result to database failed %w", err)
	}

	return nil
}
