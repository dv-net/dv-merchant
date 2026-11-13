package callback

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/processing_request"
	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/aml"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/service/receipts"
	"github.com/dv-net/dv-merchant/internal/service/store"
	"github.com/dv-net/dv-merchant/internal/service/transactions"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_receipts"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transactions"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transfer_transactions"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transfers"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_unconfirmed_transactions"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_update_balance_queue"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type ICallback interface {
	HandleDepositCallback(ctx context.Context, dto DepositWebhookDto) error
	HandleTransferCallback(ctx context.Context, dto TransferWebhookDto) error
	HandleUpdateTransferStatusCallback(ctx context.Context, dto processing_request.TransferStatusWebhook) error
}

type Service struct {
	log                            logger.Logger
	eventListener                  event.IListener
	storage                        storage.IStorage
	transactionsService            transactions.ITransaction
	unconfirmedTransactionsService transactions.IUnconfirmedTransaction
	storeService                   store.IStore
	currConvService                currconv.ICurrencyConvertor
	receiptsService                receipts.IReceiptService
}

func New(
	logger logger.Logger,
	eventListener event.IListener,
	storage storage.IStorage,
	transactionsService transactions.ITransaction,
	unconfirmedTransactionsService transactions.IUnconfirmedTransaction,
	storeService store.IStore,
	currConvService currconv.ICurrencyConvertor,
	receiptsService receipts.IReceiptService,
) ICallback {
	srv := &Service{
		log:                            logger,
		eventListener:                  eventListener,
		storage:                        storage,
		transactionsService:            transactionsService,
		unconfirmedTransactionsService: unconfirmedTransactionsService,
		storeService:                   storeService,
		currConvService:                currConvService,
		receiptsService:                receiptsService,
	}

	eventListener.Register(aml.CheckCompletedEventType, srv.handleAMLCheckCompleted)

	return srv
}

func (s *Service) HandleDepositCallback(ctx context.Context, dto DepositWebhookDto) error {
	// The receipt (or AML blocked) email must go out only after commit: the
	// transaction can roll back or be retried on a serialization failure.
	var emailEvent event.IEvent

	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		emailEvent = nil

		amount, err := decimal.NewFromString(dto.Amount)
		if err != nil {
			return fmt.Errorf("parse amount: %w", err)
		}
		fee, err := decimal.NewFromString(dto.Fee)
		if err != nil {
			return fmt.Errorf("parse fee: %w", err)
		}

		if amount.IsZero() {
			s.log.Infof("amount is zero tx_hash: %s", dto.Amount)
			return nil
		}

		_, err = s.transactionsService.GetByHashAndBcUniq(ctx, dto.Hash, dto.TxUniqKey, repos.WithTx(tx))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("transaction found error: %w", err)
		}
		if err == nil {
			s.log.Info("transaction already exists")
			return nil
		}

		if dto.ExternalWalletID == nil && dto.WalletType == models.WalletTypeHot {
			return fmt.Errorf("invalid wallet id")
		}

		wallet, err := s.storage.Wallets(repos.WithTx(tx)).GetById(ctx, *dto.ExternalWalletID)
		if err != nil {
			return fmt.Errorf("wallet found error: %w", err)
		}

		storeData, isDirty, err := s.storeService.GetStoreByWalletAddress(ctx, dto.ToAddress, dto.Currency.ID, repos.WithTx(tx))
		if err != nil {
			return fmt.Errorf("storeData found error: %w", err)
		}

		usdAmount, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
			Source:     storeData.RateSource.String(),
			From:       dto.Currency.Code,
			To:         models.CurrencyCodeUSDT,
			Amount:     amount.String(),
			StableCoin: dto.Currency.IsStablecoin,
			Scale:      &storeData.RateScale,
		})
		if err != nil {
			return fmt.Errorf("convert usd: %w", err)
		}

		if s.checkUnconfirmed(dto) {
			return s.handleUnconfirmedDeposit(ctx, dto, storeData, wallet, amount, usdAmount, isDirty, tx)
		}

		ev, err := s.handleConfirmedDeposit(ctx, dto, storeData, wallet, amount, fee, usdAmount, isDirty, tx)
		if err != nil {
			return err
		}
		if dto.IsSystem || isDirty {
			return nil
		}

		// With AML enabled the deposit check runs in this transaction. A pending
		// check holds the email back until it completes, see handleAMLCheckCompleted;
		// a check finished right away may have blocked the deposit already.
		amlPending, err := s.isAMLCheckPending(ctx, ev.Tx.ID, tx)
		if err != nil {
			return err
		}
		if amlPending {
			return nil
		}

		blocked, err := s.getBlockedTransaction(ctx, ev.Tx.ID, tx)
		if err != nil {
			return err
		}
		if blocked != nil {
			emailEvent = newDepositAMLBlockedEvent(ev.Tx, ev.Store.ID, ev.Currency, ev.WalletEmail, ev.WalletLocale, blocked)
			return nil
		}

		emailEvent = *ev
		return nil
	})
	if err != nil {
		return err
	}

	if emailEvent != nil {
		// Deposit is already committed, so a failed email must not fail the callback.
		if fireErr := s.eventListener.Fire(emailEvent); fireErr != nil {
			s.log.Errorw("eventListener fire error", "error", fireErr, "tx_hash", dto.Hash)
		}
	}

	return nil
}

func (s *Service) handleUnconfirmedDeposit(
	ctx context.Context,
	dto DepositWebhookDto,
	storeData *models.Store,
	wallet *models.Wallet,
	amount, usdAmount decimal.Decimal,
	isDirty bool,
	tx pgx.Tx,
) error {
	uTransaction, err := s.createUnconfirmedTransaction(ctx, dto, storeData, wallet, amount, usdAmount, repos.WithTx(tx))
	if err != nil {
		return fmt.Errorf("unconfirmed transaction creation: %w", err)
	}

	if dto.IsSystem || isDirty {
		return nil
	}

	if err = s.eventListener.Fire(transactions.DepositUnconfirmedEvent{
		Tx:              *uTransaction,
		Store:           *storeData,
		Currency:        *dto.Currency,
		StoreExternalID: wallet.StoreExternalID,
		WebhookEvent:    models.WebhookEventPaymentNotConfirmed,
		DBTx:            tx,
	}); err != nil {
		s.log.Errorw("eventListener fire error", "error", err)
		return fmt.Errorf("eventListener fire error: %w", err)
	}

	return nil
}

func (s *Service) handleConfirmedDeposit(
	ctx context.Context,
	dto DepositWebhookDto,
	storeData *models.Store,
	wallet *models.Wallet,
	amount, fee, usdAmount decimal.Decimal,
	isDirty bool,
	tx pgx.Tx,
) (*transactions.DepositReceiptSentEvent, error) {
	usdFee, exchangeRate, err := s.convertFeeAndExchangeRate(ctx, dto, storeData, amount, fee, usdAmount)
	if err != nil {
		return nil, err
	}

	var receiptID uuid.NullUUID
	if !dto.IsSystem {
		receipt, err := s.receiptsService.Create(ctx, repo_receipts.CreateParams{
			Status:     models.ReceiptStatusPaid,
			StoreID:    storeData.ID,
			CurrencyID: dto.Currency.ID,
			Amount:     amount,
			WalletID:   uuid.NullUUID{UUID: wallet.ID, Valid: true},
		}, repos.WithTx(tx))
		if err != nil {
			return nil, fmt.Errorf("receipt creation: %w", err)
		}
		receiptID = uuid.NullUUID{UUID: receipt.ID, Valid: true}
	}

	createParams := repo_transactions.CreateParams{
		UserID:             storeData.UserID,
		StoreID:            uuid.NullUUID{UUID: storeData.ID, Valid: true},
		ReceiptID:          receiptID,
		WalletID:           uuid.NullUUID{UUID: wallet.ID, Valid: true},
		CurrencyID:         dto.Currency.ID,
		Blockchain:         *dto.Currency.Blockchain,
		IsSystem:           dto.IsSystem,
		TxHash:             dto.Hash,
		BcUniqKey:          &dto.TxUniqKey,
		Type:               models.TransactionsTypeDeposit,
		FromAddress:        dto.FromAddress,
		ToAddress:          dto.ToAddress,
		Amount:             amount,
		AmountUsd:          decimal.NullDecimal{Decimal: usdAmount, Valid: true},
		Fee:                fee,
		WithdrawalIsManual: false,
		NetworkCreatedAt:   pgtype.Timestamp{Time: dto.NetworkCreatedAt, Valid: true},
	}

	transaction, err := s.transactionsService.CreateTransaction(ctx, createParams, repos.WithTx(tx))
	if err != nil {
		return nil, fmt.Errorf("transaction creation: %w", err)
	}

	nativeCurrency, err := dto.Currency.Blockchain.NativeCurrency()
	if err != nil {
		return nil, fmt.Errorf("get native currency: %w", err)
	}

	nativeTokenUpdateRequired := nativeCurrency == dto.Currency.ID && dto.Blockchain.RecalculateNativeBalance()
	if err = s.enqueueAddressBalanceUpdate(ctx, dto, nativeTokenUpdateRequired, repos.WithTx(tx)); err != nil {
		return nil, fmt.Errorf("can't enqueue update balance: %w", err)
	}

	if !dto.IsSystem && !isDirty {
		if err = s.eventListener.Fire(transactions.DepositReceivedEvent{
			Tx:              *transaction,
			Store:           *storeData,
			Currency:        *dto.Currency,
			StoreExternalID: wallet.StoreExternalID,
			WebhookEvent:    models.WebhookEventPaymentReceived,
			DBTx:            tx,
		}); err != nil {
			s.log.Errorw("eventListener fire error", "error", err)
			return nil, fmt.Errorf("eventListener fire error: %w", err)
		}
	}

	return &transactions.DepositReceiptSentEvent{
		Tx:              *transaction,
		Store:           *storeData,
		Currency:        *dto.Currency,
		StoreExternalID: wallet.StoreExternalID,
		WalletEmail:     resolveUserEmail(wallet),
		WalletLocale:    wallet.Locale,
		ExchangeRate:    exchangeRate,
		UsdFee:          usdFee,
	}, nil
}

func (s *Service) convertFeeAndExchangeRate(
	ctx context.Context,
	dto DepositWebhookDto,
	store *models.Store,
	amount, fee, usdAmount decimal.Decimal,
) (usdFee, exchangeRate decimal.Decimal, err error) {
	exchangeRate = decimal.NewFromFloat(usdAmount.InexactFloat64() / amount.InexactFloat64()).Round(int32(dto.Currency.Precision))

	if fee.IsZero() {
		return decimal.Zero, exchangeRate, nil
	}

	usdFee, err = s.currConvService.Convert(ctx, currconv.ConvertDTO{
		Source:     store.RateSource.String(),
		From:       dto.Currency.Code,
		To:         models.CurrencyCodeUSDT,
		Amount:     fee.String(),
		StableCoin: dto.Currency.IsStablecoin,
		Scale:      &store.RateScale,
	})
	if err != nil {
		return decimal.Zero, decimal.Zero, fmt.Errorf("convert usd fee: %w", err)
	}

	return usdFee, exchangeRate, nil
}

func (s *Service) isAMLCheckPending(ctx context.Context, txID uuid.UUID, dbTx pgx.Tx) (bool, error) {
	check, err := s.storage.AmlChecks(repos.WithTx(dbTx)).GetByTransactionID(ctx, uuid.NullUUID{UUID: txID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("fetch aml check: %w", err)
	}

	return check.Status == models.AmlCheckStatusPending, nil
}

// getBlockedTransaction returns the record of a deposit blocked by the AML rules, nil
// when the deposit is not blocked.
func (s *Service) getBlockedTransaction(ctx context.Context, txID uuid.UUID, dbTx pgx.Tx) (*models.BlockedTransaction, error) {
	blocked, err := s.storage.BlockedTransactions(repos.WithTx(dbTx)).GetByTransactionID(ctx, txID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil //nolint:nilnil
		}
		return nil, fmt.Errorf("fetch blocked transaction: %w", err)
	}

	return blocked, nil
}

func newDepositAMLBlockedEvent(
	trx models.Transaction,
	storeID uuid.UUID,
	curr models.Currency,
	walletEmail, walletLocale string,
	blocked *models.BlockedTransaction,
) transactions.DepositAMLBlockedEvent {
	return transactions.DepositAMLBlockedEvent{
		Tx:           trx,
		StoreID:      storeID,
		Currency:     curr,
		WalletEmail:  walletEmail,
		WalletLocale: walletLocale,
		Blocked:      *blocked,
	}
}

// handleAMLCheckCompleted sends the email that was held back while the AML check was
// pending: the receipt if the deposit passed, the AML blocked email if it was blocked.
func (s *Service) handleAMLCheckCompleted(ev event.IEvent) error {
	completedEv, ok := ev.(aml.CheckCompletedEvent)
	if !ok || !completedEv.Check.TransactionID.Valid {
		return nil
	}

	ctx := context.Background()

	trx, err := s.storage.Transactions().GetById(ctx, completedEv.Check.TransactionID.UUID)
	if err != nil {
		return fmt.Errorf("fetch tx: %w", err)
	}
	if trx.IsSystem || trx.Type != models.TransactionsTypeDeposit || !trx.StoreID.Valid || !trx.WalletID.Valid {
		return nil
	}

	storeData, err := s.storage.Stores().GetByID(ctx, trx.StoreID.UUID)
	if err != nil {
		return fmt.Errorf("fetch store: %w", err)
	}

	curr, err := s.storage.Currencies().GetByID(ctx, trx.CurrencyID)
	if err != nil {
		return fmt.Errorf("fetch currency: %w", err)
	}

	wallet, err := s.storage.Wallets().GetById(ctx, trx.WalletID.UUID)
	if err != nil {
		return fmt.Errorf("fetch wallet: %w", err)
	}

	// Use the rate fixed at deposit time rather than the current one.
	exchangeRate := decimal.Zero
	if !trx.Amount.IsZero() {
		exchangeRate = trx.GetAmountUsd().Div(trx.Amount).Round(int32(curr.Precision))
	}

	if completedEv.Blocked {
		blocked, err := s.getBlockedTransaction(ctx, trx.ID, nil)
		if err != nil {
			return err
		}
		if blocked == nil {
			// The verdict blocked the deposit but was not recorded: no blocked
			// transaction to refer to, and a receipt must not go out either.
			s.log.Errorw("aml blocked deposit has no blocked transaction record", "tx_id", trx.ID)
			return nil
		}

		return s.eventListener.Fire(newDepositAMLBlockedEvent(*trx, storeData.ID, *curr, resolveUserEmail(wallet), wallet.Locale, blocked))
	}

	return s.eventListener.Fire(transactions.DepositReceiptSentEvent{
		Tx:              *trx,
		Store:           *storeData,
		Currency:        *curr,
		StoreExternalID: wallet.StoreExternalID,
		WalletEmail:     resolveUserEmail(wallet),
		WalletLocale:    wallet.Locale,
		ExchangeRate:    exchangeRate,
		UsdFee:          trx.Fee.Mul(exchangeRate),
		PaymentStatus:   notify.PaymentStatusCompleted,
	})
}

func resolveUserEmail(wallet *models.Wallet) string {
	if wallet.Email.Valid {
		return wallet.Email.String
	}
	if wallet.UntrustedEmail.Valid {
		return wallet.UntrustedEmail.String
	}
	return ""
}

func (s *Service) HandleTransferCallback(ctx context.Context, dto TransferWebhookDto) error {
	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		amount, shouldContinue, err := s.validateTransferCallback(ctx, dto, tx)
		if err != nil {
			return err
		}
		if !shouldContinue {
			return nil
		}

		if s.checkUnconfirmed(dto) {
			return s.handleUnconfirmedTransfer(ctx, dto, amount, tx)
		}

		if err := s.updateTransferTxHash(ctx, dto, tx); err != nil {
			return err
		}

		return s.handleConfirmedTransfer(ctx, dto, amount, tx)
	})
}

func (s *Service) validateTransferCallback(ctx context.Context, dto TransferWebhookDto, tx pgx.Tx) (decimal.Decimal, bool, error) {
	amount, err := decimal.NewFromString(dto.Amount)
	if err != nil {
		return decimal.Zero, false, fmt.Errorf("parse amount: %w", err)
	}

	if amount.IsZero() {
		s.log.Infof("amount is zero tx_hash: %s", dto.Amount)
		return decimal.Zero, false, nil
	}

	_, err = s.transactionsService.GetByHashAndBcUniq(ctx, dto.Hash, dto.TxUniqKey, repos.WithTx(tx))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return decimal.Zero, false, fmt.Errorf("transaction found error: %w", err)
	}

	if err == nil {
		s.log.Info("transaction already exists")
		return decimal.Zero, false, nil
	}

	return amount, true, nil
}

func (s *Service) handleUnconfirmedTransfer(ctx context.Context, dto TransferWebhookDto, amount decimal.Decimal, tx pgx.Tx) error {
	switch dto.WalletType {
	case models.WalletTypeProcessing:
		return s.handleUnconfirmedProcessingTransfer(ctx, dto, amount, tx)
	case models.WalletTypeHot:
		return s.handleUnconfirmedHotWalletTransfer(ctx, dto, amount, tx)
	default:
		return fmt.Errorf("unsupported wallet type for unconfirmed transaction: %s", dto.WalletType)
	}
}

func (s *Service) handleUnconfirmedProcessingTransfer(ctx context.Context, dto TransferWebhookDto, amount decimal.Decimal, tx pgx.Tx) error {
	if !dto.TransferID.Valid {
		s.log.Infow(
			"unconfirmed withdrawal from processing transaction received without request id",
			"from", dto.FromAddress,
			"to", dto.ToAddress,
			"hash", dto.Hash,
		)
		return nil
	}

	withdrawalData, err := s.storage.WithdrawalsFromProcessing(repos.WithTx(tx)).FindByTransferID(ctx, dto.TransferID.UUID)
	if err != nil {
		return fmt.Errorf("withdrawal from processing not found: %w", err)
	}

	storeData, err := s.storeService.GetStoreByID(ctx, withdrawalData.WithdrawalFromProcessingWallet.StoreID)
	if err != nil {
		return fmt.Errorf("fetch store data: %w", err)
	}

	usdAmount, err := s.convertToUSD(ctx, amount, dto.Currency, storeData)
	if err != nil {
		return err
	}

	dummyWallet := &models.Wallet{
		ID:      uuid.Nil,
		StoreID: storeData.ID,
	}

	err = s.createUnconfirmedTransferTransaction(
		ctx, dto, storeData, dummyWallet, amount, usdAmount,
		models.TransactionsTypeWithdrawalFromProcessing,
		repos.WithTx(tx),
	)
	return err
}

func (s *Service) handleUnconfirmedHotWalletTransfer(ctx context.Context, dto TransferWebhookDto, amount decimal.Decimal, tx pgx.Tx) error {
	if dto.ExternalWalletID == nil {
		return fmt.Errorf("invalid external wallet id")
	}

	wallet, err := s.storage.Wallets(repos.WithTx(tx)).GetById(ctx, *dto.ExternalWalletID)
	if err != nil {
		return fmt.Errorf("wallet found error: %w", err)
	}

	storeData, _, err := s.storeService.GetStoreByWalletAddress(ctx, dto.FromAddress, dto.Currency.ID, repos.WithTx(tx))
	if err != nil {
		return fmt.Errorf("storeData found error: %w", err)
	}

	usdAmount, err := s.convertToUSD(ctx, amount, dto.Currency, storeData)
	if err != nil {
		return err
	}

	txType, err := s.hotWalletTransferTxType(ctx, dto, tx)
	if err != nil {
		return err
	}

	err = s.createUnconfirmedTransferTransaction(
		ctx, dto, storeData, wallet, amount, usdAmount,
		txType,
		repos.WithTx(tx),
	)
	return err
}

func (s *Service) updateTransferTxHash(ctx context.Context, dto TransferWebhookDto, tx pgx.Tx) error {
	if !dto.TransferID.Valid {
		return nil
	}

	return s.storage.Transfers(repos.WithTx(tx)).UpdateTxHash(ctx, repo_transfers.UpdateTxHashParams{
		ID:     dto.TransferID.UUID,
		TxHash: new(dto.Hash),
	})
}

func (s *Service) handleConfirmedTransfer(ctx context.Context, dto TransferWebhookDto, amount decimal.Decimal, tx pgx.Tx) error {
	switch dto.WalletType {
	case models.WalletTypeProcessing:
		return s.handleConfirmedProcessingTransfer(ctx, dto, amount, tx)
	case models.WalletTypeHot:
		return s.handleConfirmedHotWalletTransfer(ctx, dto, amount, tx)
	default:
		return fmt.Errorf("unsupported wallet type: %s", dto.WalletType)
	}
}

func (s *Service) handleConfirmedProcessingTransfer(ctx context.Context, dto TransferWebhookDto, amount decimal.Decimal, tx pgx.Tx) error {
	if !dto.TransferID.Valid {
		s.log.Infow(
			"withdrawal from processing transaction received without request id",
			"from", dto.FromAddress,
			"to", dto.ToAddress,
			"hash", dto.Hash,
		)
		return nil
	}

	withdrawalData, err := s.storage.WithdrawalsFromProcessing(repos.WithTx(tx)).FindByTransferID(ctx, dto.TransferID.UUID)
	if err != nil {
		return fmt.Errorf("withdrawal from processing not found: %w", err)
	}

	storeData, err := s.storeService.GetStoreByID(ctx, withdrawalData.WithdrawalFromProcessingWallet.StoreID)
	if err != nil {
		return fmt.Errorf("fetch store data: %w", err)
	}

	createParams, err := s.prepareCreateTxParamsByProcessingWallet(ctx, dto, storeData, amount)
	if err != nil {
		return err
	}

	createdTx, err := s.transactionsService.CreateTransaction(ctx, *createParams, repos.WithTx(tx))
	if err != nil {
		return fmt.Errorf("transaction creation: %w", err)
	}

	return s.eventListener.Fire(transactions.WithdrawalFromProcessingReceivedEvent{
		WithdrawalID: withdrawalData.WithdrawalFromProcessingWallet.ID.String(),
		Tx:           *createdTx,
		Store:        *storeData,
		Currency:     *dto.Currency,
		WebhookEvent: models.WebhookEventWithdrawalFromProcessingReceived,
		DBTx:         tx,
	})
}

func (s *Service) handleConfirmedHotWalletTransfer(ctx context.Context, dto TransferWebhookDto, amount decimal.Decimal, tx pgx.Tx) error {
	createParams, err := s.prepareCreateTxParamsByHotWallet(ctx, dto, amount, tx)
	if err != nil {
		return err
	}

	_, err = s.transactionsService.CreateTransaction(ctx, *createParams, repos.WithTx(tx))
	if err != nil {
		return fmt.Errorf("transaction creation: %w", err)
	}

	updateNativeTokenRequired := createParams.Fee.IsPositive() && dto.Blockchain.RecalculateNativeBalance()
	return s.enqueueAddressBalanceUpdate(ctx, dto, updateNativeTokenRequired, repos.WithTx(tx))
}

func (s *Service) convertToUSD(ctx context.Context, amount decimal.Decimal, currency *models.Currency, store *models.Store) (decimal.Decimal, error) {
	usdAmount, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
		Source:     store.RateSource.String(),
		From:       currency.Code,
		To:         models.CurrencyCodeUSDT,
		Amount:     amount.String(),
		StableCoin: currency.IsStablecoin,
		Scale:      &store.RateScale,
	})
	if err != nil {
		return decimal.Zero, fmt.Errorf("convert usd: %w", err)
	}
	return usdAmount, nil
}

func (s *Service) prepareCreateTxParamsByProcessingWallet(
	ctx context.Context,
	dto TransferWebhookDto,
	storeData *models.Store,
	amount decimal.Decimal,
) (*repo_transactions.CreateParams, error) {
	usdAmount, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
		Source:     storeData.RateSource.String(),
		From:       dto.Currency.Code,
		To:         "USDT",
		Amount:     amount.String(),
		StableCoin: dto.Currency.IsStablecoin,
		Scale:      &storeData.RateScale,
	})
	if err != nil {
		return nil, fmt.Errorf("convert usd: %w", err)
	}

	fee, err := decimal.NewFromString(dto.Fee)
	if err != nil {
		return nil, fmt.Errorf("parse fee: %w", err)
	}

	return &repo_transactions.CreateParams{
		UserID:             storeData.UserID,
		StoreID:            uuid.NullUUID{UUID: storeData.ID, Valid: true},
		CurrencyID:         dto.Currency.ID,
		Blockchain:         *dto.Currency.Blockchain,
		TxHash:             dto.Hash,
		BcUniqKey:          &dto.TxUniqKey,
		Type:               models.TransactionsTypeWithdrawalFromProcessing,
		FromAddress:        dto.FromAddress,
		ToAddress:          dto.ToAddress,
		Amount:             amount,
		AmountUsd:          decimal.NullDecimal{Decimal: usdAmount, Valid: true},
		Fee:                fee,
		WithdrawalIsManual: false,
		NetworkCreatedAt:   pgtype.Timestamp{Time: dto.NetworkCreatedAt, Valid: true},
	}, nil
}

func (s *Service) prepareCreateTxParamsByHotWallet(
	ctx context.Context,
	dto TransferWebhookDto,
	amount decimal.Decimal,
	tx pgx.Tx,
) (*repo_transactions.CreateParams, error) {
	if dto.ExternalWalletID == nil {
		return nil, fmt.Errorf("invalid external wallet id")
	}

	wallet, err := s.storage.Wallets(repos.WithTx(tx)).GetById(ctx, *dto.ExternalWalletID)
	if err != nil {
		return nil, fmt.Errorf("wallet found error: %w", err)
	}

	storeData, _, err := s.storeService.GetStoreByWalletAddress(ctx, dto.FromAddress, dto.Currency.ID, repos.WithTx(tx))
	if err != nil {
		return nil, fmt.Errorf("storeData found error: %w", err)
	}

	usdAmount, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
		Source:     storeData.RateSource.String(),
		From:       dto.Currency.Code,
		To:         models.CurrencyCodeUSDT,
		Amount:     amount.String(),
		StableCoin: dto.Currency.IsStablecoin,
		Scale:      &storeData.RateScale,
	})
	if err != nil {
		return nil, fmt.Errorf("convert usd: %w", err)
	}

	fee, err := decimal.NewFromString(dto.Fee)
	if err != nil {
		return nil, fmt.Errorf("parse fee: %w", err)
	}

	txType, err := s.hotWalletTransferTxType(ctx, dto, tx)
	if err != nil {
		return nil, err
	}

	return &repo_transactions.CreateParams{
		UserID:             storeData.UserID,
		StoreID:            uuid.NullUUID{UUID: storeData.ID, Valid: true},
		WalletID:           uuid.NullUUID{UUID: wallet.ID, Valid: true},
		CurrencyID:         dto.Currency.ID,
		Blockchain:         *dto.Currency.Blockchain,
		TxHash:             dto.Hash,
		BcUniqKey:          &dto.TxUniqKey,
		Type:               txType,
		FromAddress:        dto.FromAddress,
		ToAddress:          dto.ToAddress,
		Amount:             amount,
		AmountUsd:          decimal.NullDecimal{Decimal: usdAmount, Valid: true},
		Fee:                fee,
		WithdrawalIsManual: false,
		NetworkCreatedAt:   pgtype.Timestamp{Time: dto.NetworkCreatedAt, Valid: true},
	}, nil
}

func (s *Service) HandleUpdateTransferStatusCallback(ctx context.Context, dto processing_request.TransferStatusWebhook) error {
	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		step := util.TruncateString(dto.Step, 255)

		if err := s.storage.Transfers(repos.WithTx(tx)).UpdateTransferStatus(ctx, repo_transfers.UpdateTransferStatusParams{
			Status:  dto.Status,
			Stage:   models.ResolveTransferStageByStatus(dto.Status),
			Step:    new(step),
			ID:      *dto.RequestID,
			Message: dto.ErrorMessage,
		}); err != nil {
			return fmt.Errorf("update status callback: %w", err)
		}

		if err := s.syncRefundRequestStatus(ctx, dto, tx); err != nil {
			return err
		}

		if len(dto.SystemTransactions) < 1 {
			return nil
		}

		batchParams := make([]repo_transfer_transactions.BatchCreateParams, 0, len(dto.SystemTransactions))
		for _, sysTx := range dto.SystemTransactions {
			batchParams = append(batchParams, repo_transfer_transactions.BatchCreateParams{
				TransferID:        *dto.RequestID,
				TxHash:            sysTx.TxHash,
				BandwidthAmount:   sysTx.BandwidthAmount,
				EnergyAmount:      sysTx.EnergyAmount,
				NativeTokenAmount: sysTx.NativeTokenAmount,
				NativeTokenFee:    sysTx.NativeTokenFee,
				TxType:            sysTx.TxType,
				Status:            sysTx.Status,
				Step:              sysTx.Step,
			})
		}

		errChan := make(chan error, len(batchParams))
		result := s.storage.TransferTransactions(repos.WithTx(tx)).BatchCreate(ctx, batchParams)
		defer func() {
			if clsErr := result.Close(); clsErr != nil {
				s.log.Errorw("close transfer transactions batch", "error", clsErr)
			}
		}()

		wg := sync.WaitGroup{}
		wg.Add(len(batchParams))

		result.Exec(func(_ int, err error) {
			defer wg.Done()
			errChan <- err
		})

		go func() {
			wg.Wait()
			close(errChan)
		}()

		for err := range errChan {
			if err != nil {
				return fmt.Errorf("batch create transfer system transactions: %w", err)
			}
		}

		return nil
	})
}

func (s *Service) createUnconfirmedTransaction(
	ctx context.Context,
	dto DepositWebhookDto,
	store *models.Store,
	wallet *models.Wallet,
	amount decimal.Decimal,
	usdAmount decimal.Decimal,
	opts repos.Option,
) (*models.UnconfirmedTransaction, error) {
	blockchain := *dto.Currency.Blockchain
	uTransaction, err := s.unconfirmedTransactionsService.GetUnconfirmedByHash(ctx, dto.Hash, blockchain, opts)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("unconfirmed transaction found error: %w", err)
	}

	if err == nil {
		s.log.Info("unconfirmed transaction already exists")
		return uTransaction, nil
	}

	params := repo_unconfirmed_transactions.CreateParams{
		UserID:      store.UserID,
		StoreID:     uuid.NullUUID{UUID: store.ID, Valid: true},
		WalletID:    uuid.NullUUID{UUID: wallet.ID, Valid: true},
		CurrencyID:  dto.Currency.ID,
		Blockchain:  blockchain,
		TxHash:      dto.Hash,
		BcUniqKey:   &dto.TxUniqKey,
		Type:        models.TransactionsTypeDeposit,
		FromAddress: dto.FromAddress,
		ToAddress:   dto.ToAddress,
		Amount:      amount,
		AmountUsd:   decimal.NullDecimal{Decimal: usdAmount, Valid: true},
	}

	uTransaction, err = s.unconfirmedTransactionsService.CreateUnconfirmedTransaction(ctx, params, opts)
	if err != nil {
		return nil, fmt.Errorf("unconfirmed transaction creation: %w", err)
	}

	return uTransaction, nil
}

func (s *Service) createUnconfirmedTransferTransaction(
	ctx context.Context,
	dto TransferWebhookDto,
	store *models.Store,
	wallet *models.Wallet,
	amount decimal.Decimal,
	usdAmount decimal.Decimal,
	transactionType models.TransactionsType,
	opts repos.Option,
) error {
	blockchain := *dto.Currency.Blockchain
	_, err := s.unconfirmedTransactionsService.GetUnconfirmedByHash(ctx, dto.Hash, blockchain, opts)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("unconfirmed transaction found error: %w", err)
	}

	if err == nil {
		s.log.Info("unconfirmed transaction already exists")
		return nil
	}

	params := repo_unconfirmed_transactions.CreateParams{
		UserID:      store.UserID,
		StoreID:     uuid.NullUUID{UUID: store.ID, Valid: true},
		WalletID:    uuid.NullUUID{UUID: wallet.ID, Valid: true},
		CurrencyID:  dto.Currency.ID,
		Blockchain:  blockchain,
		TxHash:      dto.Hash,
		BcUniqKey:   &dto.TxUniqKey,
		Type:        transactionType,
		FromAddress: dto.FromAddress,
		ToAddress:   dto.ToAddress,
		Amount:      amount,
		AmountUsd:   decimal.NullDecimal{Decimal: usdAmount, Valid: true},
	}

	_, err = s.unconfirmedTransactionsService.CreateUnconfirmedTransaction(ctx, params, opts)
	if err != nil {
		return fmt.Errorf("unconfirmed transaction creation: %w", err)
	}

	return nil
}

func (s *Service) checkUnconfirmed(dto WebhookDtoInterface) bool {
	return dto.GetStatus() == models.TransactionStatusWaitingConfirmations ||
		dto.GetStatus() == models.TransactionStatusInMempool
}

func (s *Service) enqueueAddressBalanceUpdate(ctx context.Context, dto WebhookDtoInterface, nativeTokenUpdateRequired bool, opts repos.Option) error {
	_, err := s.storage.UpdateBalanceQueue(opts).Create(ctx, repo_update_balance_queue.CreateParams{
		CurrencyID:               dto.GetCurrency().ID,
		Address:                  dto.GetAddressForUpdateBalance(),
		NativeTokenBalanceUpdate: nativeTokenUpdateRequired,
	})
	if err != nil {
		return fmt.Errorf("enqueue address balance: %w", err)
	}

	return nil
}
