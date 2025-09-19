package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/store_request"
	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currency"
	"github.com/dv-net/dv-merchant/internal/service/exrate"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/service/transactions"
	"github.com/dv-net/dv-merchant/internal/service/wallet"
	"github.com/dv-net/dv-merchant/internal/service/webhook"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_stores"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_stores"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/dv-net/dv-merchant/pkg/rate"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type IStore interface { //nolint:interfacebloat
	GetAllStores(ctx context.Context, page int32) ([]*models.Store, error)
	GetArchivedList(ctx context.Context, userID uuid.UUID) ([]*models.Store, error)
	GetStoreByID(ctx context.Context, ID uuid.UUID) (*models.Store, error)
	GetStoresByUserID(ctx context.Context, ID uuid.UUID) ([]*models.Store, error)
	CreateStore(ctx context.Context, dto CreateStore, user *models.User, opts ...repos.Option) (*models.Store, error)
	UpdateStore(ctx context.Context, dto *store_request.UpdateRequest, ID uuid.UUID, opts ...repos.Option) (*models.Store, error)
	GetStoreByStoreAPIKey(ctx context.Context, apiKey string) (*models.Store, error)
	GetStoreByWalletAddress(ctx context.Context, address string, currencyID string, opts ...repos.Option) (*models.Store, error)
	GetStoreByWalletID(ctx context.Context, walletID uuid.UUID) (*models.Store, error)
	CheckUserHasAccess(ctx context.Context, userID, storeID uuid.UUID) (bool, error)
	GetStoreCurrencies(ctx context.Context, storeID uuid.UUID) ([]*models.Currency, error)
	PrepareTopUpDataByStore(ctx context.Context, dto wallet.CreateStoreWalletWithAddressDTO) (*TopUpData, error)
	ArchiveStore(ctx context.Context, dto ArchiveStoreDTO) error
	UnarchiveStore(ctx context.Context, dto ArchiveStoreDTO) error
}

type Service struct {
	storage             storage.IStorage
	currencyService     currency.ICurrency
	log                 logger.Logger
	webhookService      webhook.IWebHook
	eventListener       event.IListener
	exRate              exrate.IExRateSource
	rateLimiter         rate.Limiter
	rateLimitEnabled    bool
	wallets             wallet.IWalletService
	notificationService notify.INotificationService
	processingSvc       processing.IProcessingOwner
	settingSvc          setting.ISettingService
}

var _ IStore = (*Service)(nil)

const PageSize = 10

func New(
	storage storage.IStorage,
	currencyService currency.ICurrency,
	log logger.Logger,
	webhookService webhook.IWebHook,
	eventListener event.IListener,
	exRate exrate.IExRateSource,
	wallets wallet.IWalletService,
	notificationService notify.INotificationService,
	rateLimit rate.Limiter,
	rateLimitEnabled bool,
	processingSvc processing.IProcessingOwner,
	settingSvc setting.ISettingService,
) *Service {
	srv := &Service{
		storage:             storage,
		currencyService:     currencyService,
		log:                 log,
		webhookService:      webhookService,
		eventListener:       eventListener,
		exRate:              exRate,
		rateLimiter:         rateLimit,
		wallets:             wallets,
		rateLimitEnabled:    rateLimitEnabled,
		notificationService: notificationService,
		processingSvc:       processingSvc,
		settingSvc:          settingSvc,
	}
	// register event
	srv.eventListener.Register(transactions.DepositReceivedEventType, srv.handleDepositReceived)
	srv.eventListener.Register(transactions.DepositUnconfirmedEventType, srv.handleDepositReceived)
	srv.eventListener.Register(transactions.WithdrawalFromProcessingReceivedEventType, srv.handleWithdrawalReceived)

	return srv
}

func (s *Service) GetAllStores(ctx context.Context, page int32) ([]*models.Store, error) {
	if page < 1 {
		return nil, fmt.Errorf("page number must be greater than 0")
	}
	params := repo_stores.GetAllParams{
		Limit:  PageSize,
		Offset: (page - 1) * PageSize,
	}

	stores, err := s.storage.Stores().GetAll(ctx, params)
	if err != nil {
		return nil, err
	}
	return stores, nil
}

func (s *Service) GetStoreByID(ctx context.Context, id uuid.UUID) (*models.Store, error) {
	store, err := s.storage.Stores().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return store, err
}

func (s *Service) GetStoresByUserID(ctx context.Context, id uuid.UUID) ([]*models.Store, error) {
	stores, err := s.storage.Stores().GetByUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return stores, err
}

func (s *Service) GetStoreByStoreAPIKey(ctx context.Context, apiKey string) (*models.Store, error) {
	store, err := s.storage.Stores().GetStoreByStoreApiKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	return store, err
}

func (s *Service) CreateStore(ctx context.Context, dto CreateStore, user *models.User, opts ...repos.Option) (*models.Store, error) {
	params := repo_stores.CreateParams{
		Name:       dto.Name,
		UserID:     user.ID,
		CurrencyID: "USD",
		RateSource: user.RateSource,
		Status:     true,
		Site:       dto.Site,
	}
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("validate params error: %w", err)
	}

	var store *models.Store
	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		st, err := s.storage.Stores(opts...).Create(ctx, params)
		if err != nil {
			return err
		}
		// create supported store currencies
		currencies, err := s.storage.Currencies(repos.WithTx(tx)).GetCurrenciesEnabled(ctx)
		if err != nil {
			return err
		}
		for _, value := range currencies {
			err := s.CreateStoreCurrency(ctx, st, value, repos.WithTx(tx))
			if err != nil {
				return err
			}
		}
		// set user stores
		_, err = s.storage.UserStores(repos.WithTx(tx)).Create(ctx, repo_user_stores.CreateParams{
			UserID:  user.ID,
			StoreID: st.ID,
		})
		if err != nil {
			return err
		}
		// create api key
		_, err = s.CreateAPIKey(ctx, st, repos.WithTx(tx))
		if err != nil {
			return err
		}
		// create store secret
		_, err = s.GenerateSecret(ctx, st.ID, repos.WithTx(tx))
		if err != nil {
			return err
		}
		// create store settings
		err = s.CreateStoreSettings(ctx, st, repos.WithTx(tx))
		if err != nil {
			return err
		}
		store = st
		return nil
	}, opts...)
	if err != nil {
		return nil, err
	}
	return store, err
}

func (s *Service) ArchiveStore(ctx context.Context, dto ArchiveStoreDTO) error { //nolint:dupl
	if dto.User == nil || !dto.User.ProcessingOwnerID.Valid {
		return fmt.Errorf("2fa is required")
	}

	if err := s.processingSvc.ValidateTwoFactorToken(ctx, dto.User.ProcessingOwnerID.UUID, dto.OTP); err != nil {
		return ErrInvalidOTP
	}

	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		store, err := s.storage.Stores(repos.WithTx(tx)).GetByID(ctx, dto.StoreID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrStoreNotFound
		}
		if err != nil {
			return fmt.Errorf("get store by id: %w", err)
		}

		if store.UserID.String() != dto.User.ID.String() {
			return ErrUserHasNoAccess
		}

		if err := s.storage.Stores(repos.WithTx(tx)).SoftDelete(ctx, store.ID); err != nil {
			return err
		}

		removedWalletIDs, err := s.storage.Wallets(repos.WithTx(tx)).SoftDeleteByStore(ctx, store.ID)
		if err != nil {
			return err
		}

		if err = s.storage.WalletAddresses(repos.WithTx(tx)).SoftDeleteByWallets(ctx, removedWalletIDs); err != nil {
			return err
		}

		if err = s.storage.StoreAPIKeys(repos.WithTx(tx)).DisableByStore(ctx, store.ID); err != nil {
			return err
		}

		return s.storage.StoreWebhooks(repos.WithTx(tx)).DisableAllByStore(ctx, dto.StoreID)
	})
}

func (s *Service) UnarchiveStore(ctx context.Context, dto ArchiveStoreDTO) error { //nolint:dupl
	if dto.User == nil || !dto.User.ProcessingOwnerID.Valid {
		return fmt.Errorf("2fa is required")
	}

	if err := s.processingSvc.ValidateTwoFactorToken(ctx, dto.User.ProcessingOwnerID.UUID, dto.OTP); err != nil {
		return ErrInvalidOTP
	}

	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		store, err := s.storage.Stores(repos.WithTx(tx)).GetByID(ctx, dto.StoreID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrStoreNotFound
		}
		if err != nil {
			return fmt.Errorf("get store by id: %w", err)
		}

		if store.UserID.String() != dto.User.ID.String() {
			return ErrUserHasNoAccess
		}

		if err := s.storage.Stores(repos.WithTx(tx)).Restore(ctx, store.ID); err != nil {
			return err
		}

		restoredWalletIDs, err := s.storage.Wallets(repos.WithTx(tx)).RestoreByStore(ctx, store.ID)
		if err != nil {
			return err
		}

		if err = s.storage.WalletAddresses(repos.WithTx(tx)).RestoreByWallets(ctx, restoredWalletIDs); err != nil {
			return err
		}

		if err = s.storage.StoreAPIKeys(repos.WithTx(tx)).EnableByStore(ctx, store.ID); err != nil {
			return err
		}

		return s.storage.StoreWebhooks(repos.WithTx(tx)).EnableAllByStore(ctx, store.ID)
	})
}

func (s *Service) GetArchivedList(ctx context.Context, userID uuid.UUID) ([]*models.Store, error) {
	return s.storage.Stores().GetArchivedByUser(ctx, userID)
}

func (s *Service) GetStoreByWalletAddress(ctx context.Context, address string, currencyID string, opts ...repos.Option) (*models.Store, error) {
	params := repo_stores.GetStoreByWalletAddressParams{
		Address:    address,
		CurrencyID: currencyID,
	}
	store, err := s.storage.Stores(opts...).GetStoreByWalletAddress(ctx, params)
	if err != nil {
		return nil, err
	}
	return store, err
}

func (s *Service) UpdateStore(ctx context.Context, dto *store_request.UpdateRequest, id uuid.UUID, opts ...repos.Option) (*models.Store, error) {
	params := repo_stores.UpdateParams{
		Name:                     dto.Name,
		Site:                     dto.Site,
		CurrencyID:               dto.CurrencyID,
		RateSource:               models.RateSource(dto.RateSource),
		ReturnUrl:                dto.ReturnURL,
		SuccessUrl:               dto.SuccessURL,
		RateScale:                dto.RateScale,
		Status:                   dto.Status,
		MinimalPayment:           dto.MinimalPayment,
		PublicPaymentFormEnabled: dto.PublicPaymentFormEnabled,
		ID:                       id,
	}

	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("validate params error: %w", err)
	}
	store, err := s.storage.Stores(opts...).Update(ctx, params)
	if err != nil {
		return nil, err
	}
	return store, err
}

func (s *Service) CheckUserHasAccess(ctx context.Context, userID, storeID uuid.UUID) (bool, error) {
	res, err := s.storage.UserStores().CheckStoreHasUser(ctx, repo_user_stores.CheckStoreHasUserParams{
		StoreID: storeID,
		UserID:  userID,
	})
	if err != nil {
		var preparedErr error
		if !errors.Is(err, pgx.ErrNoRows) {
			preparedErr = fmt.Errorf("check user has access: %w", err)
		}
		return false, preparedErr
	}

	return res, nil
}

func (s *Service) GetStoreCurrencies(ctx context.Context, storeID uuid.UUID) ([]*models.Currency, error) {
	res, err := s.storage.Stores().GetStoreCurrencies(ctx, storeID)
	if err != nil {
		return nil, fmt.Errorf("fetch store currencies: %w", err)
	}

	return res, nil
}

func (s *Service) GetStoreByWalletID(ctx context.Context, walletID uuid.UUID) (*models.Store, error) {
	store, err := s.storage.Stores().GetStoreByWalletID(ctx, walletID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrStoreNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("fetch store by wallet ID: %w", err)
	}

	if !store.Status {
		return nil, ErrStoreDisabled
	}

	return store, nil
}
