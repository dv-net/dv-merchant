package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dv-net/dv-merchant/internal/service/aml"
	"github.com/dv-net/dv-merchant/internal/service/transactions"
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
	convertedEv, ok := ev.(transactions.TransactionEvent)
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

	ctx := context.Background()

	_, err = s.storage.StoreCurrencies().FindByStoreID(ctx, params)
	if err != nil {
		s.log.Errorw("store available currency not found", "error", err)
		return nil
	}
	if amlCheck, blocked := s.checkAMLBlock(ctx, convertedEv); blocked {
		if amlCheck != nil && amlCheck.Status != models.AmlCheckStatusPending {
			return s.sendAMLBlockedWebhook(
				ctx,
				convertedEv.GetTx(),
				convertedEv.GetStore().ID,
				convertedEv.GetCurrency(),
				convertedEv.GetStoreExternalID(),
				amlCheck,
				convertedEv.GetDatabaseTx(),
			)
		}
		return nil
	}
	return s.processWebhooksByTransactionEvent(ev, transactions.DepositReceivedEventType, payload)
}

func (s *Service) checkAMLBlock(ctx context.Context, ev transactions.TransactionEvent) (*models.AmlCheck, bool) {
	if ev.EventType() != transactions.DepositReceivedEventType {
		return nil, false
	}
	amlSettings, err := s.storage.StoreAmlSettings().GetByStoreID(ctx, ev.GetStore().ID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			s.log.Errorw("get aml settings failed", "error", err)
		}
		return nil, false
	}
	if !amlSettings.Enabled {
		return nil, false
	}
	return s.amlGate(ctx, ev, amlSettings)
}

func (s *Service) amlGate(ctx context.Context, ev transactions.TransactionEvent, settings *models.StoreAmlSetting) (*models.AmlCheck, bool) {
	amlCheck, err := s.amlService.AutoScoreDeposit(ctx, aml.AutoScoreDepositDTO{
		UserID:        ev.GetStore().UserID,
		TxID:          ev.GetTx().GetID(),
		TxHash:        ev.GetTx().GetTxHash(),
		CurrencyID:    ev.GetTx().GetCurrencyID(),
		OutputAddress: ev.GetTx().GetToAddress(),
		ProviderSlug:  settings.ProviderSlug,
		DBTx:          ev.GetDatabaseTx(),
	})
	if err != nil {
		if !errors.Is(err, aml.ErrNoProviderAvailable) {
			s.log.Errorw("aml auto check failed", "error", err)
		}
		return nil, false // fail-open
	}
	if amlCheck == nil || amlCheck.Status != models.AmlCheckStatusPending && !isScoreAboveThreshold(amlCheck.Score, settings.RiskThreshold) {
		return nil, false
	}
	if amlCheck.Status != models.AmlCheckStatusPending {
		usr, usrErr := s.storage.Users().GetByID(ctx, ev.GetStore().UserID)
		if usrErr != nil {
			s.log.Errorw("failed to get user for mark address dirty", "error", usrErr)
		} else if markErr := s.wallets.MarkAddressDirty(ctx, usr, ev.GetTx().GetToAddress()); markErr != nil {
			s.log.Errorw("failed to mark address as dirty", "error", markErr)
		}
	}
	return amlCheck, true
}

func (s *Service) handleWithdrawalReceived(ev event.IEvent) error {
	convertedEv, ok := ev.(transactions.WithdrawalFromProcessingReceivedEvent)
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

	return s.processWebhooksByTransactionEvent(ev, transactions.WithdrawalFromProcessingReceivedEventType, preparedPayload)
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

func (s *Service) processWebhooksByTransactionEvent(
	ev event.IEvent,
	eventType string,
	hookPayload []byte,
) error {
	txEv, ok := ev.(transactions.TransactionEvent)
	if !ok || txEv.EventType() != eventType {
		return nil
	}

	return s.sendWebhookForTx(
		context.Background(),
		txEv.GetTx().GetID(),
		txEv.GetStore().ID,
		txEv.GetWebhookEvent(),
		hookPayload,
		txEv.GetDatabaseTx(),
	)
}

func (s *Service) isWebhookAlreadySent(ctx context.Context, url, whType string, txID uuid.UUID, dbTx pgx.Tx) bool {
	whCheckParams := repo_webhook_send_histories.CheckWebhookWasSentParams{
		TxID: txID,
		Type: whType,
		Url:  url,
	}

	exists, err := s.storage.WebHookSendHistories(repos.WithTx(dbTx)).CheckWebhookWasSent(
		ctx,
		whCheckParams,
	)
	if err != nil {
		s.log.Errorw("check webhook status error", "error", err)
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
			Blockchain:         models.BlockchainBitcoin,
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
			Blockchain:       models.BlockchainBitcoin,
		}
	default:
		return nil, fmt.Errorf("undefined wh_type: %s", whType.String())
	}

	return txData, nil
}

func (s *Service) handleAMLCheckCompleted(ev event.IEvent) error {
	completedEv, ok := ev.(aml.CheckCompletedEvent)
	if !ok || !completedEv.Check.TransactionID.Valid {
		return nil
	}

	ctx := context.Background()
	txID := completedEv.Check.TransactionID.UUID

	tx, err := s.storage.Transactions().GetById(ctx, txID)
	if err != nil {
		return fmt.Errorf("fetch tx: %w", err)
	}

	store, err := s.storage.Stores().GetByID(ctx, tx.StoreID.UUID)
	if err != nil {
		return fmt.Errorf("fetch store: %w", err)
	}

	amlSettings, err := s.storage.StoreAmlSettings().GetByStoreID(ctx, store.ID)
	if err != nil {
		return fmt.Errorf("fetch aml settings: %w", err)
	}

	curr, err := s.storage.Currencies().GetByID(ctx, tx.CurrencyID)
	if err != nil {
		return fmt.Errorf("fetch currency: %w", err)
	}

	storeExternalID := ""
	if tx.WalletID.Valid {
		if wallet, wErr := s.storage.Wallets().GetById(ctx, tx.WalletID.UUID); wErr == nil {
			storeExternalID = wallet.StoreExternalID
		}
	}

	if isScoreAboveThreshold(completedEv.Check.Score, amlSettings.RiskThreshold) {
		s.log.Warnw("AML check score above threshold, blocking webhook", "store_id", store.ID, "tx_id", tx.ID, "score", completedEv.Check.Score, "threshold", amlSettings.RiskThreshold)
		usr, usrErr := s.storage.Users().GetByID(ctx, store.UserID)
		if usrErr != nil {
			s.log.Errorw("failed to get user for mark address dirty", "error", usrErr)
		} else if markErr := s.wallets.MarkAddressDirty(ctx, usr, tx.ToAddress); markErr != nil {
			s.log.Errorw("failed to mark address as dirty", "error", markErr)
		}
		return s.sendAMLBlockedWebhook(ctx, tx, store.ID, *curr, storeExternalID, &completedEv.Check, nil)
	}

	payload, err := s.prepareDepositHookPayload(tx, *curr, models.WebhookEventPaymentReceived, storeExternalID)
	if err != nil {
		return fmt.Errorf("prepare deposit hook payload: %w", err)
	}

	return s.sendWebhookForTx(ctx, tx.ID, store.ID, models.WebhookEventPaymentReceived, payload, nil)
}

func (s *Service) sendWebhookForTx(
	ctx context.Context,
	txID uuid.UUID,
	storeID uuid.UUID,
	whType models.WebhookEvent,
	payload []byte,
	dbTx pgx.Tx,
) error {
	webhooks, err := s.getWebhooksByStore(
		ctx,
		storeID,
		whType.String(),
		dbTx,
	)
	if err != nil {
		s.log.Errorw("store webhook not found", "error", err)
		return nil
	}
	for _, wh := range webhooks {
		if s.isWebhookAlreadySent(ctx, wh.StoreWebhook.Url, whType.String(), txID, dbTx) {
			continue
		}
		message := webhook.Message{
			TxID:      txID,
			WebhookID: wh.StoreWebhook.ID,
			Type:      whType.String(),
			Data:      payload,
			Signature: hash.SHA256Signature(payload, wh.Secret.String),
		}

		if whSendErr := s.webhookService.Send(&message, dbTx); whSendErr != nil {
			s.log.Errorw(
				"store webhook send error",
				"error", whSendErr,
				"store_id", storeID.String(),
				"tx_id", txID.String(),
				"wh_type", whType.String(),
				"wh_body", string(payload),
			)
		}
	}
	return nil
}

func (s *Service) sendAMLBlockedWebhook(
	ctx context.Context,
	tx models.ITransaction,
	storeID uuid.UUID,
	curr models.Currency,
	storeExternalID string,
	amlCheck *models.AmlCheck,
	dbTx pgx.Tx,
) error {
	payload, err := s.prepareAMLBlockedHookPayload(tx, curr, storeExternalID, amlCheck)
	if err != nil {
		return fmt.Errorf("prepare AML blocked hook payload: %w", err)
	}
	webhooks, err := s.getWebhooksByStore(
		ctx,
		storeID,
		models.WebhookEventPaymentNotConfirmed.String(), // using uncofirmed wh for no created new
		dbTx,
	)
	if err != nil {
		s.log.Errorw("store webhook not found", "error", err)
		return nil
	}

	for _, wh := range webhooks {
		if s.isWebhookAlreadySent(ctx, wh.StoreWebhook.Url, models.WebhookEventPaymentAMLBlocked.String(), tx.GetID(), dbTx) {
			continue
		}

		message := webhook.Message{
			TxID:      tx.GetID(),
			WebhookID: wh.StoreWebhook.ID,
			Type:      models.WebhookEventPaymentAMLBlocked.String(),
			Data:      payload,
			Signature: hash.SHA256Signature(payload, wh.Secret.String),
		}
		if whSendErr := s.webhookService.Send(&message, dbTx); whSendErr != nil {
			s.log.Errorw("aml blocked webhook send error", "error", whSendErr, "store_id", storeID, "tx_id", tx.GetID())
		}
	}

	return nil
}

func (s *Service) prepareAMLBlockedHookPayload(
	tx models.ITransaction,
	curr models.Currency,
	storeExternalID string,
	amlCheck *models.AmlCheck,
) ([]byte, error) {
	payload := map[string]any{
		"type":       models.WebhookEventPaymentAMLBlocked,
		"status":     models.TransactionStatusCompleted,
		"created_at": tx.GetCreatedAt(),
		"paid_at":    tx.GetNetworkCreatedAt(),
		"amount":     tx.GetAmountUsd().String(),
		"transactions": map[string]any{
			"tx_id":       tx.GetID().String(),
			"tx_hash":     tx.GetTxHash(),
			"bc_uniq_key": tx.GetBcUniqKey(),
			"created_at":  tx.GetCreatedAt(),
			"currency":    curr.Code,
			"currency_id": curr.ID,
			"blockchain":  curr.Blockchain.String(),
			"amount":      tx.GetAmount().String(),
			"amount_usd":  tx.GetAmountUsd().String(),
		},
		"wallet": map[string]any{
			"id":                tx.GetWalletID(),
			"store_external_id": storeExternalID,
		},
		"aml_check": map[string]any{
			"id":         amlCheck.ID.String(),
			"score":      amlCheck.Score,
			"status":     amlCheck.Status,
			"risk_level": amlCheck.RiskLevel,
			"created_at": amlCheck.CreatedAt,
			"updated_at": amlCheck.UpdatedAt,
		},
	}

	return json.Marshal(payload)
}
