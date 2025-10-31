package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/store_webhook_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/webhook"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_store_webhooks"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_webhook_send_histories"
	"github.com/dv-net/dv-merchant/internal/tools/hash"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type IStoreWebhooks interface {
	GetStoreWebhookByID(ctx context.Context, ID uuid.UUID) (*models.StoreWebhook, error)
	GetStoreWebhookByStoreID(ctx context.Context, storeID uuid.UUID) ([]*models.StoreWebhook, error)
	CreateStoreWebhooks(ctx context.Context, store *models.Store, dto *store_webhook_request.CreateRequest, opts ...repos.Option) (*models.StoreWebhook, error)
	UpdateStoreWebhooks(ctx context.Context, ID uuid.UUID, dto *store_webhook_request.UpdateRequest, opts ...repos.Option) (*models.StoreWebhook, error)
	DeleteStoreWebhooks(ctx context.Context, ID uuid.UUID, opts ...repos.Option) error
	SendWebhookManual(ctx context.Context, txID, userID uuid.UUID) error
	SendMockWebhook(ctx context.Context, user *models.User, whID uuid.UUID, whType models.WebhookEvent) (webhook.Result, error)
}

func (s *Service) GetStoreWebhookByID(ctx context.Context, id uuid.UUID) (*models.StoreWebhook, error) {
	storeAPIKey, err := s.storage.StoreWebhooks().GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	return storeAPIKey, nil
}

func (s *Service) CreateStoreWebhooks(ctx context.Context, store *models.Store, dto *store_webhook_request.CreateRequest, opts ...repos.Option) (*models.StoreWebhook, error) {
	params := repo_store_webhooks.CreateParams{
		Url:     dto.URL,
		StoreID: store.ID,
		Enabled: dto.Enabled,
		Events:  dto.Events,
	}
	storeWebhook, err := s.storage.StoreWebhooks(opts...).Create(ctx, params)
	if err != nil {
		return nil, err
	}
	return storeWebhook, nil
}

func (s *Service) UpdateStoreWebhooks(ctx context.Context, id uuid.UUID, dto *store_webhook_request.UpdateRequest, opts ...repos.Option) (*models.StoreWebhook, error) {
	params := repo_store_webhooks.UpdateParams{
		Url:     dto.URL,
		Enabled: dto.Enabled,
		Events:  dto.Events,
		ID:      id,
	}
	storeWebhook, err := s.storage.StoreWebhooks(opts...).Update(ctx, params)
	if err != nil {
		return nil, err
	}
	return storeWebhook, nil
}

func (s *Service) GetStoreWebhookByStoreID(ctx context.Context, storeID uuid.UUID) ([]*models.StoreWebhook, error) {
	storeWebhooks, err := s.storage.StoreWebhooks().GetByStoreId(ctx, storeID)
	if err != nil {
		return nil, err
	}
	return storeWebhooks, nil
}

func (s *Service) DeleteStoreWebhooks(ctx context.Context, id uuid.UUID, opts ...repos.Option) error {
	err := s.storage.StoreWebhooks(opts...).Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) SendWebhookManual(ctx context.Context, txID, userID uuid.UUID) error {
	transaction, err := s.storage.Transactions().GetById(ctx, txID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("fetch tx: %w", err)
	}

	tx := models.ITransaction(transaction)
	if errors.Is(err, pgx.ErrNoRows) {
		unconfirmed, fetchUnconfirmedErr := s.storage.UnconfirmedTransactions().GetById(ctx, txID)
		if fetchUnconfirmedErr != nil {
			if errors.Is(fetchUnconfirmedErr, pgx.ErrNoRows) {
				return errors.New("transaction not found")
			}

			return fmt.Errorf("fetch unconfirmed tx: %w", err)
		}

		tx = models.ITransaction(unconfirmed)
	}

	if !tx.GetType().RequiresWebhookToStore() {
		return errors.New("deposit transaction required for webhook")
	}

	hasAccess, err := s.CheckUserHasAccess(ctx, userID, tx.GetStoreID())
	if err != nil {
		return fmt.Errorf("check user access: %w", err)
	}
	if !hasAccess {
		return ErrUserHasNoAccess
	}

	curr, err := s.storage.Currencies().GetByID(ctx, tx.GetCurrencyID())
	if err != nil {
		return fmt.Errorf("fetch currency: %w", err)
	}

	storeExternalID := ""
	if tx.GetAccountID().Valid {
		wallet, fetchWalletErr := s.storage.Wallets().GetById(ctx, tx.GetAccountID().UUID)
		if fetchWalletErr != nil {
			return fmt.Errorf("fetch wallet: %w", fetchWalletErr)
		}
		storeExternalID = wallet.StoreExternalID
	}

	whType := models.WebhookEventPaymentNotConfirmed
	if tx.IsConfirmed() {
		whType = models.WebhookEventPaymentReceived
	}

	preparedPayload, err := s.prepareDepositHookPayload(tx, *curr, whType, storeExternalID)
	if err != nil {
		return fmt.Errorf("prepare hook body: %w", err)
	}

	webhooks, err := s.getWebhooksByStore(ctx, tx.GetStoreID(), whType.String(), nil)
	if err != nil {
		return fmt.Errorf("fetch store webhooks: %w", err)
	}

	for _, hook := range webhooks {
		var signature string
		if hook.Secret.Valid {
			signature = hash.SHA256Signature(preparedPayload, hook.Secret.String)
		}

		whSendErr := s.webhookService.ProcessPlainMessage(ctx, webhook.PreparedHookDto{
			ID:            uuid.NullUUID{},
			TransactionID: tx.GetID(),
			StoreID:       tx.GetStoreID(),
			IsManual:      true,
			Event:         whType.String(),
			Payload:       preparedPayload,
			Signature:     signature,
			URL:           hook.StoreWebhook.Url,
		})
		if whSendErr != nil {
			s.log.Errorw(
				"store webhook send error",
				"error", err,
				"store_id", tx.GetStoreID().String(),
				"tx_id", tx.GetID().String(),
				"tx_hash", tx.GetTxHash(),
				"wh_type", whType.String(),
				"wh_body", string(preparedPayload),
			)
		}
	}
	return nil
}

func (s *Service) getWebhooksByStore(
	ctx context.Context,
	storeID uuid.UUID,
	whType string,
	dbTx pgx.Tx,
) ([]*repo_store_webhooks.GetByStoreAndTypeRow, error) {
	params := repo_store_webhooks.GetByStoreAndTypeParams{
		StoreID:   storeID,
		EventType: whType,
	}

	webhooks, err := s.storage.StoreWebhooks(repos.WithTx(dbTx)).GetByStoreAndType(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("fetch webhooks: %w", err)
	}

	return webhooks, nil
}

func (s *Service) isWebhookAlreadySent(url, whType string, txID uuid.UUID, dbTx pgx.Tx) bool {
	whCheckParams := repo_webhook_send_histories.CheckWebhookWasSentParams{
		TxID: txID,
		Type: whType,
		Url:  url,
	}

	exists, err := s.storage.WebHookSendHistories(repos.WithTx(dbTx)).CheckWebhookWasSent(
		context.Background(),
		whCheckParams,
	)
	if err != nil {
		s.log.Errorw("check webhook status error", "error", err)
	}

	return exists
}

func (s *Service) SendMockWebhook(ctx context.Context, user *models.User, whID uuid.UUID, whType models.WebhookEvent) (webhook.Result, error) {
	wh, err := s.storage.StoreWebhooks().GetById(ctx, whID)
	if err != nil {
		return webhook.Result{}, fmt.Errorf("fetch store webhook: %w", err)
	}

	mockTxData, err := s.prepareMockTransactionDataForWhTest(whType, *wh, user.ID)
	if err != nil {
		return webhook.Result{}, fmt.Errorf("prepare mock transaction data for wh test: %w", err)
	}

	currID, err := models.BlockchainBitcoin.NativeCurrency()
	if err != nil {
		return webhook.Result{}, fmt.Errorf("get current blockchain native currency id: %w", err)
	}

	curr, err := s.storage.Currencies().GetByID(ctx, currID)
	if err != nil {
		return webhook.Result{}, fmt.Errorf("fetch currency: %w", err)
	}

	payload, err := s.prepareDepositHookPayload(mockTxData, *curr, whType, "store_external_example")
	if err != nil {
		return webhook.Result{}, fmt.Errorf("preapare payload: %w", err)
	}

	secret, err := s.storage.StoreSecrets().GetSecretByStoreID(ctx, wh.StoreID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return webhook.Result{}, fmt.Errorf("get store webhook secret: %w", err)
	}

	var sign string
	if secret != "" {
		sign = hash.SHA256Signature(payload, secret)
	}

	return s.webhookService.SendWebhook(ctx, wh.Url, payload, sign)
}
