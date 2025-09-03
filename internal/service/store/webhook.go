package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_store_currencies"
	"github.com/dv-net/dv-merchant/internal/util"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/store_webhook_request"
	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/webhook"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_store_webhooks"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_webhook_send_histories"
	"github.com/dv-net/dv-merchant/internal/tools/hash"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/goccy/go-json"
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

func (s *Service) handleDepositReceived(ev event.IEvent) error {
	convertedEv, ok := ev.(TransactionEvent)
	if !ok {
		return fmt.Errorf("invalid event type %s", ev.Type())
	}

	payload, err := s.prepareDepositHookPayload(
		convertedEv.GetTx(),
		convertedEv.GetCurrency(),
		convertedEv.GetWebhookEvent(),
		convertedEv.GetStoreExternalID(),
	)
	if err != nil {
		return fmt.Errorf("prepare deposit hook payload: %w", err)
	}

	if convertedEv.GetStore().MinimalPayment.GreaterThan(convertedEv.GetTx().GetAmountUsd()) {
		return nil
	}
	params := repo_store_currencies.FindByStoreIDParams{
		StoreID:    convertedEv.GetStore().ID,
		CurrencyID: convertedEv.GetCurrency().ID,
	}

	_, err = s.storage.StoreCurrencies().FindByStoreID(context.Background(), params)
	if err != nil {
		s.log.Error("store available currency not found %s", err)
		return nil
	}

	return s.processWebhooksByTransactionEvent(ev, DepositReceivedEventType, payload)
}

func (s *Service) handleWithdrawalReceived(ev event.IEvent) error {
	convertedEv, ok := ev.(WithdrawalFromProcessingReceivedEvent)
	if !ok {
		return fmt.Errorf("invalid event type %s", ev.Type())
	}

	payload := map[string]any{
		"type":          convertedEv.GetWebhookEvent(),
		"created_at":    convertedEv.Tx.CreatedAt,
		"paid_at":       convertedEv.Tx.NetworkCreatedAt,
		"amount":        convertedEv.Tx.GetAmountUsd().String(),
		"withdrawal_id": convertedEv.WithdrawalID,
		"transactions": map[string]any{
			"tx_id":       convertedEv.Tx.ID.String(),
			"tx_hash":     convertedEv.Tx.TxHash,
			"bc_uniq_key": convertedEv.Tx.BcUniqKey,
			"created_at":  convertedEv.Tx.CreatedAt,
			"currency":    convertedEv.Currency.Code,
			"currency_id": convertedEv.Currency.ID,
			"blockchain":  convertedEv.Currency.Blockchain.String(),
			"amount":      convertedEv.Tx.GetAmount().String(),
			"amount_usd":  convertedEv.Tx.GetAmountUsd().String(),
		},
	}

	preparedPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("prepare withdrawal hook payload: %w", err)
	}

	return s.processWebhooksByTransactionEvent(ev, WithdrawalFromProcessingReceivedEventType, preparedPayload)
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
	if tx.GetWalletID().Valid {
		wallet, fetchWalletErr := s.storage.Wallets().GetById(ctx, tx.GetWalletID().UUID)
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
			s.log.Error(
				"store webhook send error",
				err,
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

func (s *Service) processWebhooksByTransactionEvent(
	ev event.IEvent,
	eventType string,
	hookPayload []byte,
) error {
	txCreatedEvent, ok := ev.(TransactionEvent)
	if !ok || txCreatedEvent.EventType() != eventType {
		return nil
	}

	dbTx := txCreatedEvent.GetDatabaseTx()
	webhooks, err := s.getWebhooksByStore(
		context.Background(),
		txCreatedEvent.GetStore().ID,
		string(txCreatedEvent.GetWebhookEvent()),
		dbTx,
	)
	if err != nil {
		s.log.Error("store webhook not found", err)
		return nil
	}

	for _, v := range webhooks {
		if s.isWebhookAlreadySent(v.StoreWebhook.Url, string(txCreatedEvent.GetWebhookEvent()), txCreatedEvent.GetTx().GetID(), dbTx) {
			continue
		}

		if err != nil {
			s.log.Error("send webhook error", err)
			continue
		}
		message := webhook.Message{
			TxID:      txCreatedEvent.GetTx().GetID(),
			WebhookID: v.StoreWebhook.ID,
			Type:      string(txCreatedEvent.GetWebhookEvent()),
			Data:      hookPayload,
			Signature: hash.SHA256Signature(hookPayload, v.Secret.String),
		}
		whSendErr := s.webhookService.Send(&message, dbTx)

		if whSendErr != nil {
			s.log.Error(
				"store webhook send error",
				err,
				"store_id",
				txCreatedEvent.GetStore().ID.String(),
				"tx_id",
				txCreatedEvent.GetTx().GetID().String(),
				"tx_hash",
				txCreatedEvent.GetTx().GetTxHash(),
				"wh_type",
				string(txCreatedEvent.GetWebhookEvent()),
				"wh_body",
				string(hookPayload),
			)
		}
	}

	return nil
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
		s.log.Error("check webhook status error", err)
	}

	return exists
}

func (s *Service) prepareDepositHookPayload(
	tx models.ITransaction,
	curr models.Currency,
	whType models.WebhookEvent,
	storeExternalID string,
) ([]byte, error) {
	var prefix string
	if !tx.IsConfirmed() {
		prefix = "unconfirmed_"
	}
	payload := map[string]any{
		prefix + "type":       whType,
		prefix + "status":     models.TransactionStatusCompleted,
		prefix + "created_at": tx.GetCreatedAt(),
		prefix + "paid_at":    tx.GetNetworkCreatedAt(),
		prefix + "amount":     tx.GetAmountUsd().String(),
		prefix + "transactions": map[string]any{
			prefix + "tx_id":       tx.GetID().String(),
			prefix + "tx_hash":     tx.GetTxHash(),
			prefix + "bc_uniq_key": tx.GetBcUniqKey(),
			prefix + "created_at":  tx.GetCreatedAt(),
			prefix + "currency":    curr.Code,
			prefix + "currency_id": curr.ID,
			prefix + "blockchain":  curr.Blockchain.String(),
			prefix + "amount":      tx.GetAmount().String(),
			prefix + "amount_usd":  tx.GetAmountUsd().String(),
		},
		prefix + "wallet": map[string]any{
			prefix + "id":                tx.GetWalletID(),
			prefix + "store_external_id": storeExternalID,
		},
	}

	preparedPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("prepare wh body: %w", err)
	}

	return preparedPayload, nil
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

func (s *Service) prepareMockTransactionDataForWhTest(whType models.WebhookEvent, wh models.StoreWebhook, userID uuid.UUID) (models.ITransaction, error) {
	walletID, _ := uuid.NewRandom()
	preparedWalletID := uuid.NullUUID{
		UUID:  walletID,
		Valid: true,
	}
	preparedStoreID := uuid.NullUUID{
		UUID:  wh.StoreID,
		Valid: true,
	}
	txID, _ := uuid.NewRandom()
	pgTimeStamp := pgtype.Timestamp{Time: time.Now(), Valid: true}
	curr, _ := models.BlockchainBitcoin.NativeCurrency()

	var txData models.ITransaction
	switch whType {
	case models.WebhookEventPaymentReceived, models.WebhookEventWithdrawalFromProcessingReceived:
		receiptID, _ := uuid.NewRandom()
		txData = models.Transaction{
			ID:                 txID,
			UserID:             userID,
			StoreID:            preparedStoreID,
			ReceiptID:          uuid.NullUUID{UUID: receiptID, Valid: true},
			WalletID:           preparedWalletID,
			CurrencyID:         curr,
			Blockchain:         models.BlockchainBitcoin.String(),
			TxHash:             "tx_hash_example",
			BcUniqKey:          util.Pointer("bc_uniq_key_example"),
			Type:               models.TransactionsTypeDeposit,
			FromAddress:        "15muvlleOFc9nh10zTJSoM08Fil96tXBfn",
			ToAddress:          "1pmlFcSaUPBhJYeuG7ahQvTWQGWJff0IW1",
			Amount:             decimal.New(100, 0),
			AmountUsd:          decimal.NullDecimal{Decimal: decimal.New(100, 0), Valid: true},
			Fee:                decimal.Decimal{},
			WithdrawalIsManual: false,
			NetworkCreatedAt:   pgTimeStamp,
			CreatedAt:          pgTimeStamp,
			UpdatedAt:          pgTimeStamp,
			CreatedAtIndex:     pgtype.Int8{Int64: 1, Valid: true},
		}
	case models.WebhookEventPaymentNotConfirmed:
		txData = models.UnconfirmedTransaction{
			ID:               txID,
			UserID:           userID,
			StoreID:          preparedStoreID,
			WalletID:         preparedWalletID,
			CurrencyID:       curr,
			TxHash:           "tx_hash_example",
			BcUniqKey:        util.Pointer("bc_uniq_key_example"),
			Type:             models.TransactionsTypeDeposit,
			FromAddress:      "15muvlleOFc9nh10zTJSoM08Fil96tXBfn",
			ToAddress:        "1pmlFcSaUPBhJYeuG7ahQvTWQGWJff0IW1",
			Amount:           decimal.New(100, 10),
			AmountUsd:        decimal.NullDecimal{Decimal: decimal.New(100, 10), Valid: true},
			NetworkCreatedAt: pgTimeStamp,
			CreatedAt:        pgTimeStamp,
			UpdatedAt:        pgTimeStamp,
			Blockchain:       models.BlockchainBitcoin.String(),
		}
	default:
		return nil, fmt.Errorf("undefined wh_type: %s", whType.String())
	}

	return txData, nil
}
