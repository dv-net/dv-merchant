package txwebhook

import (
	"context"

	"github.com/google/uuid"

	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/aml"
	"github.com/dv-net/dv-merchant/internal/service/store"
	"github.com/dv-net/dv-merchant/internal/service/transactions"
	"github.com/dv-net/dv-merchant/internal/service/webhook"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

type IService interface {
	SendWebhookManual(ctx context.Context, txID, userID uuid.UUID) error
	SendMockWebhook(ctx context.Context, user *models.User, whID uuid.UUID, whType models.WebhookEvent) (webhook.Result, error)
}

var _ IService = (*Service)(nil)

type Service struct {
	storage        storage.IStorage
	log            logger.Logger
	webhookService webhook.IWebHook
	amlService     aml.IService
	storeService   store.IStore
}

func New(
	storage storage.IStorage,
	log logger.Logger,
	webhookService webhook.IWebHook,
	amlService aml.IService,
	eventListener event.IListener,
	storeService store.IStore,
) *Service {
	srv := &Service{
		storage:        storage,
		log:            log,
		webhookService: webhookService,
		amlService:     amlService,
		storeService:   storeService,
	}

	eventListener.Register(transactions.DepositReceivedEventType, srv.handleDepositReceived)
	eventListener.Register(transactions.DepositUnconfirmedEventType, srv.handleDepositReceived)
	eventListener.Register(transactions.WithdrawalFromProcessingReceivedEventType, srv.handleWithdrawalReceived)
	eventListener.Register(aml.CheckCompletedEventType, srv.handleAMLCheckCompleted)

	return srv
}
