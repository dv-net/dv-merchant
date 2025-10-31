package callback

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dv-net/dv-merchant/internal/constant"
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/processing_request"
	"github.com/dv-net/dv-merchant/internal/event"
	eventtypes "github.com/dv-net/dv-merchant/internal/event/types"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/invoice"
	"github.com/dv-net/dv-merchant/internal/service/receipts"
	"github.com/dv-net/dv-merchant/internal/service/store"
	"github.com/dv-net/dv-merchant/internal/service/transactions"
	"github.com/dv-net/dv-merchant/internal/service/wallet"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_invoices"
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

type walletHandler struct {
	walletType constant.WalletAddressType
	handler    func(*Service, context.Context, DepositWebhookDto, pgx.Tx) error
}

var walletHandlers = []walletHandler{
	{constant.WalletAddress, (*Service).handleDepositForStaticWallet},
	{constant.RotateAddress, (*Service).handleDepositForRotateWallet},
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
	walletService                  wallet.IWalletService
	invoiceService                 invoice.IInvoiceService
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
	wService wallet.IWalletService,
	iService invoice.IInvoiceService,
) ICallback {
	return &Service{
		log:                            logger,
		eventListener:                  eventListener,
		storage:                        storage,
		transactionsService:            transactionsService,
		unconfirmedTransactionsService: unconfirmedTransactionsService,
		storeService:                   storeService,
		currConvService:                currConvService,
		receiptsService:                receiptsService,
		walletService:                  wService,
		invoiceService:                 iService,
	}
}

func (s *Service) HandleDepositCallback(ctx context.Context, dto DepositWebhookDto) error { //nolint:funlen
	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		_, err := s.transactionsService.GetByHashAndBcUniq(ctx, dto.Hash, dto.TxUniqKey, repos.WithTx(tx))
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

		for _, wh := range walletHandlers {
			walletAddress, err := s.walletService.GetByAccountID(ctx, *dto.ExternalWalletID, wh.walletType)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("wallet address found error: %w", err)
			}

			if len(walletAddress) != 0 {
				err := wh.handler(s, ctx, dto, tx)
				if err != nil {
					s.log.Errorw("handle deposit error", "error", err)
				}
			}
		}
		return nil
	})
}

func (s *Service) handleDepositForStaticWallet(ctx context.Context, dto DepositWebhookDto, tx pgx.Tx) error {
	storeData, err := s.storeService.GetStoreByWalletAddress(ctx, dto.ToAddress, dto.Currency.ID, repos.WithTx(tx))
	if err != nil {
		return fmt.Errorf("storeData found error: %w", err)
	}

	calculation, err := s.calculateAmountAndFee(ctx, dto, storeData, nil)
	if err != nil {
		return err
	}

	staticWallet, err := s.storage.Wallets(repos.WithTx(tx)).GetById(ctx, *dto.ExternalWalletID)
	if err != nil {
		return fmt.Errorf("wallet found error: %w", err)
	}

	// Either take untrusted email or trusted email. Trusted email has priority.
	var userEmail string
	if staticWallet.UntrustedEmail.Valid {
		userEmail = staticWallet.UntrustedEmail.String
	}
	if staticWallet.Email.Valid {
		userEmail = staticWallet.Email.String
	}

	if s.checkUnconfirmed(dto) {
		uTransaction, err := s.createUnconfirmedTransaction(ctx, dto, storeData, staticWallet.ID, uuid.NullUUID{}, calculation.Amount, calculation.UsdAmount, repos.WithTx(tx))
		if err != nil {
			return fmt.Errorf("unconfirmed transaction creation: %w", err)
		}

		if dto.IsSystem {
			return nil
		}

		err = s.eventListener.Fire(transactions.DepositUnconfirmedEvent{
			Tx:              *uTransaction,
			Store:           *storeData,
			Currency:        *dto.Currency,
			StoreExternalID: staticWallet.StoreExternalID,
			WebhookEvent:    models.WebhookEventPaymentNotConfirmed,
			DBTx:            tx,
		})
		if err != nil {
			s.log.Errorw("eventListener fire error", "error", err)
			return fmt.Errorf("eventListener fire error: %w", err)
		}
		return nil
	}
	var receiptID uuid.NullUUID

	if !dto.IsSystem {
		receipt, err := s.receiptsService.Create(ctx, repo_receipts.CreateParams{
			Status:     models.ReceiptStatusPaid,
			StoreID:    storeData.ID,
			CurrencyID: dto.Currency.ID,
			Amount:     calculation.Amount,
			AccountID:  uuid.NullUUID{UUID: staticWallet.ID, Valid: true},
		}, repos.WithTx(tx))
		if err != nil {
			return fmt.Errorf("receipt creation: %w", err)
		}
		receiptID = uuid.NullUUID{UUID: receipt.ID, Valid: true}
	}

	createPrams := repo_transactions.CreateParams{
		UserID:             storeData.UserID,
		StoreID:            uuid.NullUUID{UUID: storeData.ID, Valid: true},
		ReceiptID:          receiptID,
		AccountID:          uuid.NullUUID{UUID: staticWallet.ID, Valid: true},
		CurrencyID:         dto.Currency.ID,
		Blockchain:         dto.Currency.Blockchain.String(),
		IsSystem:           dto.IsSystem,
		TxHash:             dto.Hash,
		BcUniqKey:          &dto.TxUniqKey,
		Type:               models.TransactionsTypeDeposit,
		FromAddress:        dto.FromAddress,
		ToAddress:          dto.ToAddress,
		Amount:             calculation.Amount,
		AmountUsd:          decimal.NullDecimal{Decimal: calculation.UsdAmount, Valid: true},
		Fee:                calculation.Fee,
		WithdrawalIsManual: false,
		NetworkCreatedAt:   pgtype.Timestamp{Time: dto.NetworkCreatedAt, Valid: true},
	}

	transaction, err := s.transactionsService.CreateTransaction(ctx, createPrams, repos.WithTx(tx))
	if err != nil {
		return fmt.Errorf("transaction creation: %w", err)
	}

	nativeCurrency, err := dto.Currency.Blockchain.NativeCurrency()
	if err != nil {
		return fmt.Errorf("get native currency: %w", err)
	}

	nativeTokenUpdateRequired := nativeCurrency == dto.Currency.ID && dto.Blockchain.RecalculateNativeBalance()
	if err = s.enqueueAddressBalanceUpdate(ctx, dto, nativeTokenUpdateRequired, repos.WithTx(tx)); err != nil {
		return fmt.Errorf("cant't enqueue update balance: %w", err)
	}

	// if system transaction no send webhook
	if dto.IsSystem {
		return nil
	}
	// event for webhook
	err = s.eventListener.Fire(transactions.DepositReceivedEvent{
		Tx:              *transaction,
		Store:           *storeData,
		Currency:        *dto.Currency,
		StoreExternalID: staticWallet.StoreExternalID,
		WebhookEvent:    models.WebhookEventPaymentReceived,
		DBTx:            tx,
	})

	if !dto.IsSystem {
		err = s.eventListener.Fire(transactions.DepositReceiptSentEvent{
			Tx:              *transaction,
			Store:           *storeData,
			Currency:        *dto.Currency,
			StoreExternalID: staticWallet.StoreExternalID,
			DBTx:            tx,
			WalletEmail:     userEmail,
			WalletLocale:    staticWallet.Locale,
			ExchangeRate:    calculation.ExchangeRate,
			UsdFee:          calculation.UsdFee,
		})
	}
	if err != nil {
		s.log.Errorw("eventListener fire error", "error", err)
		return fmt.Errorf("eventListener fire error: %w", err)
	}

	return nil
}

func (s *Service) handleDepositForRotateWallet(ctx context.Context, dto DepositWebhookDto, tx pgx.Tx) error {
	storeData, err := s.storeService.GetStoreByWalletAddress(ctx, dto.ToAddress, dto.Currency.ID, repos.WithTx(tx))
	if err != nil {
		return err
	}

	wAddress, err := s.walletService.GetByAccountAndCurrency(ctx, *dto.ExternalWalletID, dto.Currency.ID)
	if err != nil {
		return err
	}

	invoiceInfo, err := s.storage.InvoiceAddresses(repos.WithTx(tx)).GetByWalletAddressID(ctx, wAddress.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	hasInvoice := false
	if err == nil {
		if invoiceInfo.Status != constant.InvoiceStatusExpired && invoiceInfo.Status != constant.InvoiceStatusCancelled {
			hasInvoice = true
		}
	}

	var rate *decimal.Decimal
	if hasInvoice && invoiceInfo.RateAtCreation.Valid {
		rate = &invoiceInfo.RateAtCreation.Decimal
	}

	calculation, err := s.calculateAmountAndFee(ctx, dto, storeData, rate)
	if err != nil {
		return err
	}

	if s.checkUnconfirmed(dto) {
		var invoiceID uuid.NullUUID
		if hasInvoice {
			invoiceID = uuid.NullUUID{UUID: invoiceInfo.InvoiceID, Valid: true}
		}

		_, err := s.createUnconfirmedTransaction(
			ctx,
			dto,
			storeData,
			wAddress.ID,
			invoiceID,
			calculation.Amount,
			calculation.UsdAmount,
			repos.WithTx(tx),
		)

		if err != nil {
			return fmt.Errorf("unconfirmed transaction creation: %w", err)
		}

		if dto.IsSystem {
			return nil
		}

		if hasInvoice {
			err = s.invoiceService.UpdateStatus(ctx, invoiceInfo.InvoiceID, constant.InvoiceStatusWaiting, repos.WithTx(tx))
			if err != nil {
				return fmt.Errorf("invoice update status: %w", err)
			}
		}
		// todo event for webhook

		return nil
	}

	var receiptID uuid.NullUUID

	if !dto.IsSystem {
		receipt, err := s.receiptsService.Create(ctx, repo_receipts.CreateParams{
			Status:     models.ReceiptStatusPaid,
			StoreID:    storeData.ID,
			CurrencyID: dto.Currency.ID,
			Amount:     calculation.Amount,
			AccountID:  wAddress.AccountID,
		}, repos.WithTx(tx))
		if err != nil {
			return fmt.Errorf("receipt creation: %w", err)
		}
		receiptID = uuid.NullUUID{UUID: receipt.ID, Valid: true}
	}

	var invoiceID uuid.NullUUID
	if hasInvoice {
		invoiceID = uuid.NullUUID{UUID: invoiceInfo.InvoiceID, Valid: true}
	}

	createPrams := repo_transactions.CreateParams{
		UserID:             storeData.UserID,
		StoreID:            uuid.NullUUID{UUID: storeData.ID, Valid: true},
		ReceiptID:          receiptID,
		AccountID:          wAddress.AccountID,
		InvoiceID:          invoiceID,
		CurrencyID:         dto.Currency.ID,
		Blockchain:         dto.Currency.Blockchain.String(),
		IsSystem:           dto.IsSystem,
		TxHash:             dto.Hash,
		BcUniqKey:          &dto.TxUniqKey,
		Type:               models.TransactionsTypeDeposit,
		FromAddress:        dto.FromAddress,
		ToAddress:          dto.ToAddress,
		Amount:             calculation.Amount,
		AmountUsd:          decimal.NullDecimal{Decimal: calculation.UsdAmount, Valid: true},
		Fee:                calculation.Fee,
		WithdrawalIsManual: false,
		NetworkCreatedAt:   pgtype.Timestamp{Time: dto.NetworkCreatedAt, Valid: true},
	}

	_, err = s.transactionsService.CreateTransaction(ctx, createPrams, repos.WithTx(tx))
	if err != nil {
		return fmt.Errorf("transaction creation: %w", err)
	}

	nativeCurrency, err := dto.Currency.Blockchain.NativeCurrency()
	if err != nil {
		return fmt.Errorf("get native currency: %w", err)
	}

	nativeTokenUpdateRequired := nativeCurrency == dto.Currency.ID && dto.Blockchain.RecalculateNativeBalance()
	if err = s.enqueueAddressBalanceUpdate(ctx, dto, nativeTokenUpdateRequired, repos.WithTx(tx)); err != nil {
		return fmt.Errorf("cant't enqueue update balance: %w", err)
	}

	if hasInvoice {
		sumInvoice, err := s.storage.Transactions(repos.WithTx(tx)).CalculateTotalAmountsByInvoice(ctx, uuid.NullUUID{UUID: invoiceInfo.InvoiceID, Valid: true})
		if err != nil {
			return fmt.Errorf("calculate total amount by invoice: %w", err)
		}
		inv, err := s.storage.Invoices(repos.WithTx(tx)).UpdateReceivedAmount(ctx, repo_invoices.UpdateReceivedAmountParams{
			ID:                invoiceInfo.InvoiceID,
			ReceivedAmountUsd: sumInvoice.TotalAmountUsd,
		})
		if err != nil {
			return fmt.Errorf("failed to update invoice amounts: %w", err)
		}
		newStatus := s.determineInvoiceStatus(inv)
		if newStatus != inv.Status {
			uInv, err := s.storage.Invoices(repos.WithTx(tx)).UpdateStatus(ctx, repo_invoices.UpdateStatusParams{
				ID:     inv.ID,
				Status: newStatus,
			})
			if err != nil {
				return fmt.Errorf("failed to update invoice status: %w", err)
			}

			if newStatus.CanReleaseWallet() {
				err = s.walletService.ReleaseByAccountID(ctx, wAddress.AccountID.UUID, repos.WithTx(tx))
				if err != nil {
					return fmt.Errorf("failed to release wallet: %w", err)
				}
			}
			txs, err := s.storage.Transactions(repos.WithTx(tx)).GetByInvoiceID(ctx, uuid.NullUUID{UUID: inv.ID, Valid: true})
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("failed to get transaction by invoice: %w", err)
			}

			err = s.eventListener.Fire(eventtypes.ChangeInvoiceStatusEvent{
				Store:        storeData,
				Invoice:      uInv,
				Transactions: txs,
				WebhookEvent: models.WebhookEventInvoiceChangeStatus,
				DBTx:         tx,
			})
			if err != nil {
				return fmt.Errorf("failed to fire event: %w", err)
			}
		}
	}

	return nil
}

func (s *Service) HandleTransferCallback(ctx context.Context, dto TransferWebhookDto) error {
	// skip unconfirmed transaction
	if dto.Status == models.TransactionStatusWaitingConfirmations || dto.Status == models.TransactionStatusInMempool {
		return nil
	}

	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		amount, err := decimal.NewFromString(dto.Amount)
		if err != nil {
			return fmt.Errorf("parse amount: %w", err)
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

		if dto.TransferID.Valid {
			if err = s.storage.Transfers(repos.WithTx(tx)).UpdateTxHash(ctx, repo_transfers.UpdateTxHashParams{
				ID:     dto.TransferID.UUID,
				TxHash: util.Pointer(dto.Hash),
			}); err != nil {
				return fmt.Errorf("update transfer tx hash: %w", err)
			}
		}

		switch dto.WalletType {
		case models.WalletTypeProcessing:
			if !dto.TransferID.Valid {
				s.log.Infow(
					"withdrawal from processing transaction received without request id",
					"from",
					dto.FromAddress,
					"to",
					dto.ToAddress,
					"hash",
					dto.Hash,
				)

				return nil
			}

			withdrawalData, err := s.storage.WithdrawalsFromProcessing(repos.WithTx(tx)).FindByTransferID(
				ctx,
				dto.TransferID.UUID,
			)
			if err != nil {
				return fmt.Errorf("withdrawal from processing not found")
			}

			storeData, err := s.storeService.GetStoreByID(ctx, withdrawalData.WithdrawalFromProcessingWallet.StoreID)
			if err != nil {
				return fmt.Errorf("fetch store data: %w", err)
			}

			createParams, prepareTxErr := s.prepareCreateTxParamsByProcessingWallet(
				ctx,
				dto,
				storeData,
				amount,
			)
			if prepareTxErr != nil {
				return prepareTxErr
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
		case models.WalletTypeHot:
			createParams, prepareTxErr := s.prepareCreateTxParamsByHotWallet(ctx, dto, amount, tx)
			if prepareTxErr != nil {
				return prepareTxErr
			}

			_, err = s.transactionsService.CreateTransaction(ctx, *createParams, repos.WithTx(tx))
			if err != nil {
				return fmt.Errorf("transaction creation: %w", err)
			}

			// update balance native token if need
			updateNativeTokenRequired := createParams.Fee.IsPositive() && dto.Blockchain.RecalculateNativeBalance()
			if err = s.enqueueAddressBalanceUpdate(ctx, dto, updateNativeTokenRequired, repos.WithTx(tx)); err != nil {
				return fmt.Errorf("cant't update balance: %w", err)
			}
		default:
			return fmt.Errorf("unsupported wallet type: %s", dto.WalletType)
		}
		return nil
	})
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

	return &repo_transactions.CreateParams{
		UserID:             storeData.UserID,
		StoreID:            uuid.NullUUID{UUID: storeData.ID, Valid: true},
		CurrencyID:         dto.Currency.ID,
		Blockchain:         dto.Currency.Blockchain.String(),
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

	storeData, err := s.storeService.GetStoreByWalletAddress(ctx, dto.FromAddress, dto.Currency.ID, repos.WithTx(tx))
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

	return &repo_transactions.CreateParams{
		UserID:             storeData.UserID,
		StoreID:            uuid.NullUUID{UUID: storeData.ID, Valid: true},
		AccountID:          uuid.NullUUID{UUID: wallet.ID, Valid: true},
		CurrencyID:         dto.Currency.ID,
		Blockchain:         dto.Currency.Blockchain.String(),
		TxHash:             dto.Hash,
		BcUniqKey:          &dto.TxUniqKey,
		Type:               models.TransactionsTypeTransfer,
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
		if err := s.storage.Transfers(repos.WithTx(tx)).UpdateTransferStatus(ctx, repo_transfers.UpdateTransferStatusParams{
			Status:  dto.Status,
			Stage:   models.ResolveTransferStageByStatus(dto.Status),
			Step:    util.Pointer(dto.Step),
			ID:      *dto.RequestID,
			Message: dto.ErrorMessage,
		}); err != nil {
			return fmt.Errorf("update status callback: %w", err)
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
	walletID uuid.UUID,
	invoiceID uuid.NullUUID,
	amount decimal.Decimal,
	usdAmount decimal.Decimal,
	opts repos.Option,
) (*models.UnconfirmedTransaction, error) {
	blockchain := string(*dto.Currency.Blockchain)
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
		AccountID:   uuid.NullUUID{UUID: walletID, Valid: true},
		CurrencyID:  dto.Currency.ID,
		Blockchain:  blockchain,
		TxHash:      dto.Hash,
		BcUniqKey:   &dto.TxUniqKey,
		Type:        models.TransactionsTypeDeposit,
		FromAddress: dto.FromAddress,
		ToAddress:   dto.ToAddress,
		Amount:      amount,
		AmountUsd:   decimal.NullDecimal{Decimal: usdAmount, Valid: true},
		InvoiceID:   invoiceID,
	}

	uTransaction, err = s.unconfirmedTransactionsService.CreateUnconfirmedTransaction(ctx, params, opts)
	if err != nil {
		return nil, fmt.Errorf("unconfirmed transaction creation: %w", err)
	}

	return uTransaction, nil
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

func (s *Service) calculateAmountAndFee(ctx context.Context, dto WebhookDtoInterface, storeData *models.Store, rate *decimal.Decimal) (*AmountFeeCalculation, error) {
	amount, err := decimal.NewFromString(dto.GetAmount())
	if err != nil {
		return nil, fmt.Errorf("parse amount: %w", err)
	}
	fee, err := decimal.NewFromString(dto.GetFee())
	if err != nil {
		return nil, fmt.Errorf("parse fee: %w", err)
	}

	if amount.IsZero() {
		s.log.Infof("amount is zero tx_hash: %s", dto.GetAmount())
		return nil, fmt.Errorf("amount is zero tx_hash: %s", dto.GetAmount())
	}

	currency := dto.GetCurrency()

	// If rate is provided (fixed at invoice creation), use it directly
	if rate != nil {
		usdAmount := amount.Mul(*rate)
		usdFee := decimal.Zero
		if !fee.IsZero() {
			usdFee = fee.Mul(*rate)
		}

		return &AmountFeeCalculation{
			Amount:       amount,
			Fee:          fee,
			UsdAmount:    usdAmount,
			UsdFee:       usdFee,
			ExchangeRate: *rate,
		}, nil
	}

	// Calculate USD amount
	usdAmount, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
		Source:     storeData.RateSource.String(),
		From:       currency.Code,
		To:         models.CurrencyCodeUSDT,
		Amount:     amount.String(),
		StableCoin: currency.IsStablecoin,
		Scale:      &storeData.RateScale,
	})
	if err != nil {
		return nil, fmt.Errorf("convert usd: %w", err)
	}

	// Calculate exchange rate
	exchangeRate := decimal.NewFromFloat(usdAmount.InexactFloat64() / amount.InexactFloat64()).Round(int32(currency.Precision))

	// Calculate USD fee
	usdFee := decimal.Zero
	if !fee.IsZero() {
		usdFee, err = s.currConvService.Convert(ctx, currconv.ConvertDTO{
			Source:     storeData.RateSource.String(),
			From:       currency.Code,
			To:         models.CurrencyCodeUSDT,
			Amount:     fee.String(),
			StableCoin: currency.IsStablecoin,
			Scale:      &storeData.RateScale,
		})
		if err != nil {
			return nil, fmt.Errorf("convert usd fee: %w", err)
		}
	}

	return &AmountFeeCalculation{
		Amount:       amount,
		Fee:          fee,
		UsdAmount:    usdAmount,
		UsdFee:       usdFee,
		ExchangeRate: exchangeRate,
	}, nil
}

// determineInvoiceStatus determines the invoice status based on received and expected amounts
func (s *Service) determineInvoiceStatus(invoice *models.Invoice) constant.InvoiceStatus {
	received := invoice.ReceivedAmountUsd
	expected := invoice.ExpectedAmountUsd
	now := time.Now()
	expiresAt, _ := invoice.ExpiresAt.Value()

	if expiresAt != nil {
		expiryTime, ok := expiresAt.(time.Time)
		if ok && expiryTime.Before(now) && received.IsZero() {
			return constant.InvoiceStatusExpired
		}
	}

	tolerance := expected.Mul(decimal.NewFromFloat(0.01))
	upperBound := expected.Add(tolerance)

	if received.GreaterThanOrEqual(expected) && received.LessThanOrEqual(upperBound) {
		return constant.InvoiceStausPaid
	}

	if received.GreaterThan(upperBound) {
		return constant.InvoiceStatusOverpaid
	}

	if received.GreaterThan(decimal.Zero) && received.LessThan(expected) {
		return constant.InvoiceStatusUnderpaid
	}

	return constant.InvoiceStatusPending
}
