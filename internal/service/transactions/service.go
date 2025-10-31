package transactions

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/transactions_request"
	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/eproxy"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transactions"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallets"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/go-mods/excel"
	"github.com/gocarina/gocsv"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xuri/excelize/v2"
)

type ITransaction interface {
	GetByID(ctx context.Context, ID uuid.UUID) (*models.Transaction, error)
	GetByHashAndBcUniq(ctx context.Context, hash string, bcUniqKey string, opts ...repos.Option) (*models.Transaction, error)
	CreateTransaction(ctx context.Context, params repo_transactions.CreateParams, opts ...repos.Option) (*models.Transaction, error)
	GetStoreTransactions(ctx context.Context, storeID uuid.UUID, page int32) ([]*models.Transaction, error)
	GetUserTransactions(ctx context.Context, userID uuid.UUID, params GetUserTransactionsDTO) (*storecmn.FindResponseWithFullPagination[*repo_transactions.FindRow], error)
	GetTransactionStats(ctx context.Context, user *models.User, params transactions_request.GetStatistics) ([]*repo_transactions.StatisticsRow, error)
	DownloadUserTransactions(ctx context.Context, userID uuid.UUID, params transactions_request.GetByUserExported) (*bytes.Buffer, error)
	DepositStatistics(ctx context.Context, params StatisticsParams) ([]StatisticsDTO, error)
	GetTransactionInfo(ctx context.Context, userID uuid.UUID, hash string) (TransactionInfoDto, error)
}

type IWalletTransaction interface {
	GetLastWalletDepositTransactions(ctx context.Context, walletID uuid.UUID) ([]ShortTransactionInfo, error)
	GetWalletInfoWithTransactionsByAddress(ctx context.Context, userID uuid.UUID, address string) ([]WalletWithTransactionsInfo, error)
}

type Service struct {
	storage             storage.IStorage
	conv                currconv.ICurrencyConvertor
	epr                 eproxy.IExplorerProxy
	log                 logger.Logger
	eventListener       event.IListener
	notificationService notify.INotificationService
}

const PageSize = 10

var (
	_ IWalletTransaction = (*Service)(nil)
	_ ITransaction       = (*Service)(nil)
)

func New(
	log logger.Logger,
	storage storage.IStorage,
	epr eproxy.IExplorerProxy,
	conv currconv.ICurrencyConvertor,
	eventListener event.IListener,
	notificationService notify.INotificationService,
) *Service {
	srv := &Service{
		storage:             storage,
		epr:                 epr,
		conv:                conv,
		log:                 log,
		eventListener:       eventListener,
		notificationService: notificationService,
	}

	srv.eventListener.Register(DepositReceiptSentEventType, srv.handleDepositReceiptSent)

	return srv
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	transaction, err := s.storage.Transactions().GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (s *Service) GetTransactionInfo(ctx context.Context, userID uuid.UUID, hash string) (TransactionInfoDto, error) {
	res, err := s.storage.Transactions().FindTransactionByHashAndUserID(
		ctx,
		repo_transactions.FindTransactionByHashAndUserIDParams{
			TxHash: hash,
			UserID: userID,
		},
	)
	if err != nil {
		preparedErr := fmt.Errorf("fetch tx: %w", err)
		if errors.Is(err, pgx.ErrNoRows) {
			preparedErr = ErrTransactionNotFound
		}

		return TransactionInfoDto{}, preparedErr
	}

	var nwCreatedAt *time.Time
	if res.NetworkCreatedAt.Valid {
		nwCreatedAt = &res.NetworkCreatedAt.Time
	}
	var walletUpdated *time.Time
	if res.WalletUpdatedAt.Valid {
		walletUpdated = &res.WalletUpdatedAt.Time
	}
	var receiptID *uuid.UUID
	if res.ReceiptID.Valid {
		receiptID = &res.ReceiptID.UUID
	}

	preparedResult := TransactionInfoDto{
		ID:          res.ID,
		IsConfirmed: res.IsConfirmed,
		UserID:      res.UserID,
		StoreID:     &res.StoreID.UUID,
		ReceiptID:   receiptID,
		Wallet: TransactionWalletInfoDto{
			ID:              res.AccountID.UUID,
			WalletStoreID:   res.WalletStoreID,
			StoreExternalID: res.StoreExternalID,
			WalletCreatedAt: res.WalletCreatedAt.Time,
			WalletUpdatedAt: walletUpdated,
		},
		CurrencyID:       res.CurrencyID,
		Blockchain:       res.Blockchain,
		TxHash:           res.TxHash,
		Type:             res.Type,
		FromAddress:      res.FromAddress.String,
		ToAddress:        res.ToAddress,
		Amount:           res.Amount,
		AmountUsd:        &res.AmountUsd.Decimal,
		Fee:              res.Fee,
		NetworkCreatedAt: nwCreatedAt,
		CreatedAt:        &res.CreatedAt.Time,
	}

	webhookHistories, err := s.storage.WebHookSendHistories().GetAllByTxID(ctx, res.ID)
	if err != nil {
		return preparedResult, fmt.Errorf("get webhook histories: %w", err)
	}

	whSendData := make([]TransactionWhHistoryDto, 0, len(webhookHistories))
	for _, val := range webhookHistories {
		var storeUUID uuid.UUID
		if val.StoreID.Valid {
			storeUUID = val.StoreID.UUID
		}
		txHistory := TransactionWhHistoryDto{
			ID:                 val.ID,
			StoreID:            storeUUID,
			WhType:             val.Type,
			URL:                val.Url,
			WhStatus:           val.Status,
			Request:            val.Request,
			ResponseStatusCode: val.ResponseStatusCode,
			CreatedAt:          &val.CreatedAt.Time,
		}
		if val.Response != nil {
			txHistory.Response = val.Response
		}

		whSendData = append(whSendData, txHistory)
	}

	preparedResult.WebhookHistory = whSendData
	return preparedResult, nil
}

func (s *Service) GetByHashAndBcUniq(ctx context.Context, hash string, bcUniqKey string, opts ...repos.Option) (*models.Transaction, error) {
	params := repo_transactions.GetTransactionByHashAndBcUniqKeyParams{
		BcUniqKey: &bcUniqKey,
		TxHash:    hash,
	}
	transaction, err := s.storage.Transactions(opts...).GetTransactionByHashAndBcUniqKey(ctx, params)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (s *Service) CreateTransaction(ctx context.Context, params repo_transactions.CreateParams, opts ...repos.Option) (*models.Transaction, error) {
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("validate params error: %w", err)
	}

	transaction, err := s.storage.Transactions(opts...).Create(ctx, params)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (s *Service) GetStoreTransactions(ctx context.Context, storeID uuid.UUID, page int32) ([]*models.Transaction, error) {
	if page < 1 {
		return nil, fmt.Errorf("page number must be greater than 0")
	}

	params := repo_transactions.GetTransactionsByStoreIdParams{
		StoreID: uuid.NullUUID{
			UUID:  storeID,
			Valid: true,
		},
		Limit:  PageSize,
		Offset: (page - 1) * PageSize,
	}

	transactions, err := s.storage.Transactions().GetTransactionsByStoreId(ctx, params)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (s *Service) GetUserTransactions(
	ctx context.Context,
	userID uuid.UUID,
	dto GetUserTransactionsDTO,
) (*storecmn.FindResponseWithFullPagination[*repo_transactions.FindRow], error) {
	var dateFrom *time.Time
	if dto.DateFrom != nil {
		date, err := util.ParseDate(*dto.DateFrom)
		if err != nil {
			return nil, err
		}

		dateFrom = date
	}

	var dateTo *time.Time
	if dto.DateTo != nil {
		date, err := util.ParseDate(*dto.DateTo)
		if err != nil {
			return nil, err
		}

		dateTo = date
	}

	return s.storage.Transactions().GetByUser(ctx, repo_transactions.GetByUserParams{
		UserID:           userID,
		Currencies:       dto.Currencies,
		StoreUUIDs:       dto.StoreUuids,
		WalletAddress:    dto.WalletAddress,
		ToAddresses:      dto.ToAddress,
		FromAddresses:    dto.FromAddress,
		Type:             dto.Type,
		IsSystem:         dto.IsSystem,
		MinAmount:        dto.MinAmountUSD,
		Blockchain:       dto.Blockchain,
		DateFrom:         dateFrom,
		DateTo:           dateTo,
		CommonFindParams: *dto.CommonParams,
	})
}

func (s *Service) DownloadUserTransactions(ctx context.Context, userID uuid.UUID, params transactions_request.GetByUserExported) (*bytes.Buffer, error) {
	commonParams := storecmn.NewCommonFindParams()
	if params.PageSize != nil {
		commonParams.PageSize = params.PageSize
	}
	if params.Page != nil {
		commonParams.Page = params.Page
	}

	var dateFrom *time.Time
	if params.DateFrom != nil {
		date, err := util.ParseDate(*params.DateFrom)
		if err != nil {
			return nil, err
		}

		dateFrom = date
	}

	var dateTo *time.Time
	if params.DateTo != nil {
		date, err := util.ParseDate(*params.DateFrom)
		if err != nil {
			return nil, err
		}

		dateTo = date
	}

	userTransactions, err := s.storage.Transactions().GetByUser(ctx, repo_transactions.GetByUserParams{
		UserID:           userID,
		Currencies:       params.Currencies,
		StoreUUIDs:       params.StoreUuids,
		WalletAddress:    params.WalletAddress,
		ToAddresses:      params.ToAddress,
		FromAddresses:    params.FromAddress,
		Type:             params.Type,
		IsSystem:         params.IsSystem,
		MinAmount:        params.MinAmountUSD,
		Blockchain:       params.Blockchain,
		DateFrom:         dateFrom,
		DateTo:           dateTo,
		CommonFindParams: *commonParams,
	})
	if err != nil {
		return nil, fmt.Errorf("fetch transactions: %w", err)
	}

	transactions, err := s.prepareTransactionsForExport(ctx, userTransactions.Items)
	if err != nil {
		return nil, fmt.Errorf("prepare transactions for export: %w", err)
	}
	txBuffer := new(bytes.Buffer)
	switch params.Format {
	case "csv":
		if err := gocsv.Marshal(transactions, txBuffer); err != nil {
			return nil, fmt.Errorf("marshal transactions: %w", err)
		}
	case "xlsx":
		excelFile := excelize.NewFile()
		if err := excelFile.SetSheetName(excelFile.GetSheetName(excelFile.GetActiveSheetIndex()), "User transactions"); err != nil {
			return nil, fmt.Errorf("set active sheet name: %w", err)
		}
		defer func() { _ = excelFile.Close() }()
		excelWriter, err := excel.NewWriter(excelFile)
		if err != nil {
			return nil, fmt.Errorf("create excel writer: %w", err)
		}
		if err := excelWriter.SetActiveSheetName("User transactions"); err != nil {
			return nil, fmt.Errorf("set active sheet name: %w", err)
		}
		if err := excelWriter.Marshal(&transactions); err != nil {
			return nil, fmt.Errorf("marshal transactions: %w", err)
		}
		if _, err := excelWriter.File.WriteTo(txBuffer); err != nil {
			return nil, fmt.Errorf("write to buffer: %w", err)
		}
	}
	return txBuffer, nil
}

func (s *Service) prepareTransactionsForExport(ctx context.Context, txs []*repo_transactions.FindRow) ([]*UserTransactionModel, error) {
	transactions := make([]*UserTransactionModel, 0, len(txs))
	for _, tx := range txs {
		store, err := s.storage.Stores().GetByID(ctx, tx.StoreID.UUID)
		if err != nil {
			return nil, fmt.Errorf("get store by id: %w", err)
		}
		userTxModel := &UserTransactionModel{
			StoreName:        store.Name,
			CurrencyID:       tx.CurrencyID,
			Blockchain:       tx.Blockchain,
			TxHash:           tx.TxHash,
			Type:             tx.Type.String(),
			FromAddress:      tx.FromAddress,
			ToAddress:        tx.ToAddress,
			Amount:           tx.Amount.String(),
			Name:             tx.Currency.Name,
			CreatedAt:        tx.CreatedAt.Time,
			NetworkCreatedAt: tx.NetworkCreatedAt.Time,
			Fee:              tx.Fee.String(),
			UserEmail:        tx.UserEmail,
			IsSystem:         tx.IsSystem,
		}
		if tx.ReceiptID.Valid {
			userTxModel.ReceiptID = tx.ReceiptID.UUID.String()
		}
		if tx.AmountUsd.Valid {
			userTxModel.AmountUsd = tx.AmountUsd.Decimal.String()
		}
		transactions = append(transactions, userTxModel)
	}
	return transactions, nil
}

func (s *Service) GetLastWalletDepositTransactions(ctx context.Context, walletID uuid.UUID) ([]ShortTransactionInfo, error) {
	res, err := s.storage.Transactions().FindLastWalletTransactions(ctx, repo_transactions.FindLastWalletTransactionsParams{
		AccountID: uuid.NullUUID{UUID: walletID, Valid: true},
		Type:      models.TransactionsTypeDeposit,
		Limit:     10,
	})
	if err != nil {
		return nil, fmt.Errorf("find last transactions: %w", err)
	}

	preparedRes := make([]ShortTransactionInfo, 0, len(res))
	for _, val := range res {
		amountUsd := ""
		if val.AmountUsd.Valid {
			amountUsd = val.AmountUsd.Decimal.String()
		}
		preparedRes = append(preparedRes, ShortTransactionInfo{
			IsConfirmed:  val.IsConfirmed,
			CurrencyCode: val.CurrCode,
			Hash:         val.TxHash,
			Amount:       val.Amount.String(),
			AmountUSD:    amountUsd,
			Type:         val.Type,
			CreatedAt:    val.CreatedAt.Time,
		})
	}

	return preparedRes, nil
}

func (s *Service) GetWalletInfoWithTransactionsByAddress(
	ctx context.Context,
	userID uuid.UUID,
	address string,
) ([]WalletWithTransactionsInfo, error) {
	res, err := s.storage.Wallets().GetWalletWithStore(ctx, repo_wallets.GetWalletWithStoreParams{
		Address: address,
		UserID:  userID,
	})
	if err != nil {
		return nil, fmt.Errorf("fetch wallets with stores: %w", err)
	}

	result := make(map[string]WalletWithTransactionsInfo, len(res))
	for _, walletData := range res {
		preparedWalletData, ok := result[walletData.Address]
		if !ok {
			preparedWalletData = WalletWithTransactionsInfo{
				Address:         walletData.Address,
				StoreUUID:       walletData.Store.ID,
				WalletID:        walletData.WalletID,
				StoreExternalID: walletData.StoreExternalID,
			}
		}

		if slices.Index(preparedWalletData.Currencies, walletData.CurrencyID) == -1 {
			preparedWalletData.Currencies = append(preparedWalletData.Currencies, walletData.CurrencyID)
		}

		result[walletData.Address] = preparedWalletData
	}

	preparedRes := make([]WalletWithTransactionsInfo, 0, len(result))
	for walletAddr, data := range result {
		transactions, fetchTxErr := s.prepareTxInfoByWallet(ctx, data.WalletID, walletAddr)
		if fetchTxErr != nil {
			return nil, fmt.Errorf("fetch tx data: %w", fetchTxErr)
		}

		data.Transactions = append(data.Transactions, transactions...)
		preparedRes = append(preparedRes, data)
	}

	return preparedRes, nil
}

func (s *Service) prepareTxInfoByWallet(ctx context.Context, walletID uuid.UUID, address string) ([]WalletTransactionInfo, error) {
	transactions, err := s.storage.Transactions().GetWalletTransactions(ctx, repo_transactions.GetWalletTransactionsParams{
		AccountID: uuid.NullUUID{
			UUID:  walletID,
			Valid: true,
		},
		ToAddress: address,
	})
	if err != nil {
		return nil, fmt.Errorf("fetch tx: %w", err)
	}

	preparedData := make([]WalletTransactionInfo, 0, len(transactions))
	for _, val := range transactions {
		preparedData = append(preparedData, WalletTransactionInfo{
			CurrencyID: val.CurrencyID,
			Hash:       val.TxHash,
			From:       val.FromAddress,
			To:         val.ToAddress,
			CreatedAt:  val.NetworkCreatedAt.Time,
		})
	}

	return preparedData, nil
}
