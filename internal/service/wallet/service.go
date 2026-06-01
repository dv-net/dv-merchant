package wallet

import (
	"bytes"
	"context"
	"sync"
	"sync/atomic"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/currency"
	"github.com/dv-net/dv-merchant/internal/service/eproxy"
	"github.com/dv-net/dv-merchant/internal/service/exrate"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// IWalletReader — read wallet data
type IWalletReader interface {
	GetWallet(ctx context.Context, ID uuid.UUID) (*models.Wallet, error)
	GetFullDataByID(ctx context.Context, ID uuid.UUID) (*GetAllByStoreIDResponse, error)
	GetWalletsInfo(ctx context.Context, usr *models.User, address string) ([]*WithBlockchains, error)
	SummarizeUserWalletsByCurrency(ctx context.Context, userID uuid.UUID, rates *exrate.Rates, minBalance decimal.Decimal) ([]SummaryDTO, error)
	UpdateLocale(ctx context.Context, walletID uuid.UUID, locale string, opts ...repos.Option) error
}

// IWalletWriter — create/update data
type IWalletWriter interface {
	StoreWalletWithAddress(ctx context.Context, dto CreateStoreWalletWithAddressDTO, amount string) (*WithAddressDto, error)
}

// IWalletAddressManager — address manager
type IWalletAddressManager interface {
	MarkAddressDirty(ctx context.Context, usr *models.User, address string) error
	RefreshWalletAddress(ctx context.Context, walletID uuid.UUID, address string) error
	LoadPrivateAddresses(ctx context.Context, dto LoadPrivateKeyDTO) (*bytes.Buffer, error)
}

// IWalletNotifier — notification
type IWalletNotifier interface {
	SendUserWalletNotification(ctx context.Context, walletID uuid.UUID, selectCurrency *string) error
}

// IWalletStatistics — statistic
type IWalletStatistics interface {
	FetchTronResourceStatistics(ctx context.Context, user *models.User, dto FetchTronStatisticsParams) (map[string]CombinedStats, error)
}

// IWalletService — main interface
type IWalletService interface {
	IWalletReader
	IWalletWriter
	IWalletAddressManager
	IWalletNotifier
	IWalletStatistics
}

type Service struct {
	cfg               *config.Config
	storage           storage.IStorage
	logger            logger.Logger
	currencyService   currency.ICurrency
	processingService processing.IProcessingWallet
	exrateService     exrate.IExRateSource
	currConvService   currconv.ICurrencyConvertor
	settingService    setting.ISettingService
	eproxyService     eproxy.IExplorerProxy
	notification      notify.INotificationService

	updateBalanceInProgress         atomic.Bool
	updateProcessingStatsInProgress atomic.Bool

	muMap sync.Map
}

var _ IWalletService = (*Service)(nil)
var _ IWalletBalances = (*Service)(nil)

func New(
	cfg *config.Config,
	storage storage.IStorage,
	logger logger.Logger,
	currencyService currency.ICurrency,
	processingService processing.IProcessingWallet,
	exrateService exrate.IExRateSource,
	currConvService currconv.ICurrencyConvertor,
	settingsService setting.ISettingService,
	eproxyService eproxy.IExplorerProxy,
	notification notify.INotificationService,
) *Service {
	return &Service{
		cfg:               cfg,
		storage:           storage,
		logger:            logger,
		currencyService:   currencyService,
		processingService: processingService,
		exrateService:     exrateService,
		currConvService:   currConvService,
		settingService:    settingsService,
		eproxyService:     eproxyService,
		notification:      notification,
	}
}
