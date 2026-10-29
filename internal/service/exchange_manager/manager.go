package exchange_manager

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/exchange/binance"
	"github.com/dv-net/dv-merchant/internal/service/exchange/bingx"
	"github.com/dv-net/dv-merchant/internal/service/exchange/bitget"
	"github.com/dv-net/dv-merchant/internal/service/exchange/bybit"
	"github.com/dv-net/dv-merchant/internal/service/exchange/gateio"
	"github.com/dv-net/dv-merchant/internal/service/exchange/htx"
	"github.com/dv-net/dv-merchant/internal/service/exchange/kraken"
	"github.com/dv-net/dv-merchant/internal/service/exchange/kucoin"
	"github.com/dv-net/dv-merchant/internal/service/exchange/mexc"
	"github.com/dv-net/dv-merchant/internal/service/exchange/okx"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_user_keys"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_exchanges"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/utils"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

const exchangeHTTPTimeout = 30 * time.Second

type IExchangeClient interface { //nolint:interfacebloat
	TestConnection(ctx context.Context) error
	GetAccountBalance(ctx context.Context) ([]*models.AccountBalanceDTO, error)
	GetCurrencyBalance(ctx context.Context, currency string) (*decimal.Decimal, error)
	GetExchangeSymbols(ctx context.Context) ([]*models.ExchangeSymbolDTO, error)
	GetDepositAddresses(ctx context.Context, currency, network string) ([]*models.DepositAddressDTO, error)
	CreateWithdrawalOrder(ctx context.Context, args *models.CreateWithdrawalOrderParams) (*models.ExchangeWithdrawalDTO, error)
	CreateSpotOrder(ctx context.Context, from string, to string, side string, ticker string, amount *decimal.Decimal, rule *models.OrderRulesDTO) (*models.ExchangeOrderDTO, error)
	GetOrderRule(ctx context.Context, ticker string) (*models.OrderRulesDTO, error)
	GetOrderRules(ctx context.Context, tickers ...string) ([]*models.OrderRulesDTO, error)
	GetOrderDetails(ctx context.Context, args *models.GetOrderByIDParams) (*models.OrderDetailsDTO, error)
	GetWithdrawalRules(ctx context.Context, ccys ...string) ([]*models.WithdrawalRulesDTO, error)
	GetWithdrawalByID(ctx context.Context, args *models.GetWithdrawalByIDParams) (*models.WithdrawalStatusDTO, error)
	GetConnectionHash() string
}

type IExchangeManager interface {
	GetPublicDriver(ctx context.Context, slug models.ExchangeSlug) (IExchangeClient, error)
	GetDefaultDriver(ctx context.Context, userID uuid.UUID) (IExchangeClient, error)
	GetDriver(ctx context.Context, slug models.ExchangeSlug, userID uuid.UUID) (IExchangeClient, error)
	GetDriverWithNetConfig(ctx context.Context, slug models.ExchangeSlug, userID uuid.UUID, netCfg utils.NetConfig) (IExchangeClient, error)
	CreateDriver(ctx context.Context, slug models.ExchangeSlug, apiKey, secretKey, passphrase string, netCfg utils.NetConfig) (IExchangeClient, error)
	GetNetConfig(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) (utils.NetConfig, error)
}

type Manager struct {
	l               logger.Logger
	storage         storage.IStorage
	store           limiter.Store
	currConvService currconv.ICurrencyConvertor
}

func NewManager(l logger.Logger, storage storage.IStorage, currConvService currconv.ICurrencyConvertor) IExchangeManager {
	return &Manager{
		l:               l,
		storage:         storage,
		store:           memory.NewStore(),
		currConvService: currConvService,
	}
}

func (o *Manager) CreateDriver(ctx context.Context, slug models.ExchangeSlug, apiKey, secretKey, passphrase string, netCfg utils.NetConfig) (IExchangeClient, error) {
	if !slug.Valid() {
		return nil, fmt.Errorf("invalid exchange slug provided")
	}
	httpClient, err := newHTTPClient(netCfg)
	if err != nil {
		return nil, err
	}
	switch slug {
	case models.ExchangeSlugBinance:
		return o.createBinanceServiceRaw(ctx, apiKey, secretKey, httpClient)
	case models.ExchangeSlugHtx:
		return o.createHtxServiceRaw(ctx, apiKey, secretKey, httpClient)
	case models.ExchangeSlugOkx:
		return o.createOkxServiceRaw(ctx, apiKey, secretKey, passphrase, httpClient)
	case models.ExchangeSlugBitget:
		return o.createBitgetServiceRaw(ctx, apiKey, secretKey, passphrase, httpClient)
	case models.ExchangeSlugKucoin:
		return o.createKucoinServiceRaw(ctx, apiKey, secretKey, passphrase, httpClient)
	case models.ExchangeSlugBybit:
		return o.createBybitServiceRaw(ctx, apiKey, secretKey, httpClient)
	case models.ExchangeSlugGateio:
		return o.createGateioServiceRaw(ctx, apiKey, secretKey, httpClient)
	case models.ExchangeSlugBingx:
		return o.createBingxServiceRaw(ctx, apiKey, secretKey, httpClient)
	case models.ExchangeSlugMexc:
		return o.createMexcServiceRaw(ctx, apiKey, secretKey, httpClient)
	case models.ExchangeSlugKraken:
		return o.createKrakenServiceRaw(ctx, apiKey, secretKey, httpClient)
	}
	return nil, fmt.Errorf("slug %s does not exists", slug.String())
}

func (o *Manager) GetPublicDriver(ctx context.Context, slug models.ExchangeSlug) (IExchangeClient, error) {
	switch slug {
	case models.ExchangeSlugHtx:
		return o.createPublicHtxService(ctx)
	case models.ExchangeSlugOkx:
		return nil, fmt.Errorf("okx public client is not supported")
	case models.ExchangeSlugBinance:
		return o.createPublicBinanceService(ctx)
	case models.ExchangeSlugBitget:
		return o.createPublicBitgetService(ctx)
	case models.ExchangeSlugKucoin:
		return o.createPublicKucoinService(ctx)
	case models.ExchangeSlugBybit:
		return o.createPublicBybitService(ctx)
	case models.ExchangeSlugGateio:
		return o.createPublicGateioService(ctx)
	case models.ExchangeSlugBingx:
		return nil, fmt.Errorf("bingx public client is not supported")
	case models.ExchangeSlugMexc:
		return nil, fmt.Errorf("mexc public client is not supported")
	case models.ExchangeSlugKraken:
		return o.createKrakenServiceRaw(ctx, "-", "-", nil)
	default:
		return nil, fmt.Errorf("slug %s does not exists", slug.String())
	}
}

func (o *Manager) GetDefaultDriver(ctx context.Context, userID uuid.UUID) (IExchangeClient, error) {
	usr, err := o.storage.Users().GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if usr.ExchangeSlug == nil {
		return nil, fmt.Errorf("user has no exchanges setup")
	}
	if !usr.ExchangeSlug.Valid() {
		return nil, fmt.Errorf("user is missing current driver")
	}

	return o.GetDriver(ctx, *usr.ExchangeSlug, usr.ID)
}

func (o *Manager) GetDriver(ctx context.Context, slug models.ExchangeSlug, userID uuid.UUID) (IExchangeClient, error) {
	netCfg, err := o.GetNetConfig(ctx, userID, slug)
	if err != nil {
		return nil, err
	}
	return o.GetDriverWithNetConfig(ctx, slug, userID, netCfg)
}

// GetDriverWithNetConfig creates a driver with the user's saved keys but the given network settings.
// Used to check new proxy / interface settings before saving them.
func (o *Manager) GetDriverWithNetConfig(ctx context.Context, slug models.ExchangeSlug, userID uuid.UUID, netCfg utils.NetConfig) (IExchangeClient, error) {
	if !slug.Valid() {
		return nil, fmt.Errorf("invalid exchange slug provided")
	}
	httpClient, err := newHTTPClient(netCfg)
	if err != nil {
		return nil, err
	}

	switch slug {
	case models.ExchangeSlugHtx:
		return o.createHtxService(ctx, userID, httpClient)
	case models.ExchangeSlugOkx:
		return o.createOkxService(ctx, userID, httpClient)
	case models.ExchangeSlugBinance:
		return o.createBinanceService(ctx, userID, httpClient)
	case models.ExchangeSlugBitget:
		return o.createBitgetService(ctx, userID, httpClient)
	case models.ExchangeSlugKucoin:
		return o.createKucoinService(ctx, userID, httpClient)
	case models.ExchangeSlugGateio:
		return o.createGateService(ctx, userID, httpClient)
	case models.ExchangeSlugBybit:
		return o.createBybitService(ctx, userID, httpClient)
	case models.ExchangeSlugBingx:
		return o.createBingxService(ctx, userID, httpClient)
	case models.ExchangeSlugMexc:
		return o.createMexcService(ctx, userID, httpClient)
	case models.ExchangeSlugKraken:
		return o.createKrakenService(ctx, userID, httpClient)
	default:
		return nil, fmt.Errorf("slug %s does not exists", slug.String())
	}
}

func (o *Manager) createOkxServiceRaw(ctx context.Context, apiKey, secretKey, passphrase string, httpClient *http.Client) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugOkx)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return okx.NewService(o.l, apiKey, secretKey, passphrase, baseURL, o.storage, o.store, o.currConvService, httpClient)
}

func (o *Manager) createOkxService(ctx context.Context, userID uuid.UUID, httpClient *http.Client) (IExchangeClient, error) {
	keys, err := o.storage.ExchangeUserKeys().GetKeysByExchangeSlug(ctx, repo_exchange_user_keys.GetKeysByExchangeSlugParams{
		UserID:       userID,
		ExchangeSlug: models.ExchangeSlugOkx,
	})
	if err != nil {
		return nil, err
	}

	keyMap := make(map[string]string)
	for _, key := range keys {
		keyMap[string(key.Name)] = key.Value
	}

	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugOkx)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}
	apiKey := keyMap[models.ExchangeKeyNameAPIKey.String()]
	secretKey := keyMap[models.ExchangeKeyNameSecretKey.String()]
	passphrase := keyMap[models.ExchangeKeyNamePassPhrase.String()]

	return okx.NewService(o.l, apiKey, secretKey, passphrase, baseURL, o.storage, o.store, o.currConvService, httpClient)
}

func (o *Manager) createHtxServiceRaw(ctx context.Context, accessKey, secretKey string, httpClient *http.Client) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugHtx)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return htx.NewService(o.l, accessKey, secretKey, baseURL, o.storage, o.store, o.currConvService, httpClient)
}

func (o *Manager) createHtxService(ctx context.Context, userID uuid.UUID, httpClient *http.Client) (IExchangeClient, error) { //nolint:dupl
	keys, err := o.storage.ExchangeUserKeys().GetKeysByExchangeSlug(ctx, repo_exchange_user_keys.GetKeysByExchangeSlugParams{
		UserID:       userID,
		ExchangeSlug: models.ExchangeSlugHtx,
	})
	if err != nil {
		return nil, err
	}

	keyMap := make(map[string]string)
	for _, key := range keys {
		keyMap[string(key.Name)] = key.Value
	}

	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugHtx)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}
	accessKey := keyMap[models.ExchangeKeyNameAccessKey.String()]
	secretKey := keyMap[models.ExchangeKeyNameSecretKey.String()]

	return htx.NewService(o.l, accessKey, secretKey, baseURL, o.storage, o.store, o.currConvService, httpClient)
}

func (o *Manager) createPublicHtxService(ctx context.Context) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugHtx)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return htx.NewService(o.l, "-", "-", baseURL, o.storage, o.store, o.currConvService, nil)
}

func (o *Manager) createBinanceServiceRaw(ctx context.Context, apiKey, secretKey string, httpClient *http.Client) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugBinance)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return binance.NewService(o.l, apiKey, secretKey, false, baseURL, o.storage, o.currConvService, httpClient)
}

func (o *Manager) createBinanceService(ctx context.Context, userID uuid.UUID, httpClient *http.Client) (IExchangeClient, error) {
	keys, err := o.storage.ExchangeUserKeys().GetKeysByExchangeSlug(ctx, repo_exchange_user_keys.GetKeysByExchangeSlugParams{
		UserID:       userID,
		ExchangeSlug: models.ExchangeSlugBinance,
	})
	if err != nil {
		return nil, err
	}

	keyMap := make(map[string]string)
	for _, key := range keys {
		keyMap[string(key.Name)] = key.Value
	}

	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugBinance)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}
	accessKey := keyMap[models.ExchangeKeyNameAPIKey.String()]
	secretKey := keyMap[models.ExchangeKeyNameSecretKey.String()]

	return binance.NewService(o.l, accessKey, secretKey, false, baseURL, o.storage, o.currConvService, httpClient)
}

func (o *Manager) createPublicBinanceService(ctx context.Context) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugBinance)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return binance.NewService(o.l, "-", "-", true, baseURL, o.storage, o.currConvService, nil)
}

func (o *Manager) createBitgetServiceRaw(ctx context.Context, accessKey, secretKey, passPhrase string, httpClient *http.Client) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugBitget)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return bitget.NewService(o.l, accessKey, secretKey, passPhrase, false, baseURL, o.storage, o.store, o.currConvService, httpClient)
}

func (o *Manager) createBitgetService(ctx context.Context, userID uuid.UUID, httpClient *http.Client) (IExchangeClient, error) { //nolint:dupl
	keys, err := o.storage.ExchangeUserKeys().GetKeysByExchangeSlug(ctx, repo_exchange_user_keys.GetKeysByExchangeSlugParams{
		UserID:       userID,
		ExchangeSlug: models.ExchangeSlugBitget,
	})
	if err != nil {
		return nil, err
	}

	keyMap := make(map[string]string)
	for _, key := range keys {
		keyMap[string(key.Name)] = key.Value
	}

	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugBitget)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}
	accessKey := keyMap[models.ExchangeKeyNameAccessKey.String()]
	secretKey := keyMap[models.ExchangeKeyNameSecretKey.String()]
	passPhrase := keyMap[models.ExchangeKeyNamePassPhrase.String()]

	return bitget.NewService(o.l, accessKey, secretKey, passPhrase, false, baseURL, o.storage, o.store, o.currConvService, httpClient)
}

func (o *Manager) createPublicBitgetService(ctx context.Context) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugBitget)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return bitget.NewService(o.l, "-", "-", "-", true, baseURL, o.storage, o.store, o.currConvService, nil)
}

func (o *Manager) createKucoinServiceRaw(ctx context.Context, accessKey, secretKey, passPhrase string, httpClient *http.Client) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugKucoin)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return kucoin.NewService(o.l, accessKey, secretKey, passPhrase, false, baseURL, o.storage, o.store, o.currConvService, httpClient)
}

func (o *Manager) createKucoinService(ctx context.Context, userID uuid.UUID, httpClient *http.Client) (IExchangeClient, error) { //nolint:dupl
	keys, err := o.storage.ExchangeUserKeys().GetKeysByExchangeSlug(ctx, repo_exchange_user_keys.GetKeysByExchangeSlugParams{
		UserID:       userID,
		ExchangeSlug: models.ExchangeSlugKucoin,
	})
	if err != nil {
		return nil, err
	}

	keyMap := make(map[string]string)
	for _, key := range keys {
		keyMap[string(key.Name)] = key.Value
	}

	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugKucoin)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}
	accessKey := keyMap[models.ExchangeKeyNameAccessKey.String()]
	secretKey := keyMap[models.ExchangeKeyNameSecretKey.String()]
	passPhrase := keyMap[models.ExchangeKeyNamePassPhrase.String()]

	return kucoin.NewService(o.l, accessKey, secretKey, passPhrase, false, baseURL, o.storage, o.store, o.currConvService, httpClient)
}

func (o *Manager) createPublicKucoinService(ctx context.Context) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugKucoin)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return kucoin.NewService(o.l, "-", "-", "-", true, baseURL, o.storage, o.store, o.currConvService, nil)
}

func (o *Manager) createBybitServiceRaw(ctx context.Context, apiKey, secretKey string, httpClient *http.Client) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugBybit)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return bybit.NewService(o.l, apiKey, secretKey, baseURL, o.storage, o.store, o.currConvService, httpClient)
}

func (o *Manager) createBybitService(ctx context.Context, userID uuid.UUID, httpClient *http.Client) (IExchangeClient, error) { //nolint:dupl
	keys, err := o.storage.ExchangeUserKeys().GetKeysByExchangeSlug(ctx, repo_exchange_user_keys.GetKeysByExchangeSlugParams{
		UserID:       userID,
		ExchangeSlug: models.ExchangeSlugBybit,
	})
	if err != nil {
		return nil, err
	}

	keyMap := make(map[string]string)
	for _, key := range keys {
		keyMap[string(key.Name)] = key.Value
	}

	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugBybit)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}
	apiKey := keyMap[models.ExchangeKeyNameAccessKey.String()]
	secretKey := keyMap[models.ExchangeKeyNameSecretKey.String()]

	return bybit.NewService(o.l, apiKey, secretKey, baseURL, o.storage, o.store, o.currConvService, httpClient)
}

func (o *Manager) createPublicBybitService(ctx context.Context) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugBybit)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return bybit.NewService(o.l, "-", "-", baseURL, o.storage, o.store, o.currConvService, nil)
}

func (o *Manager) createGateService(ctx context.Context, userID uuid.UUID, httpClient *http.Client) (IExchangeClient, error) { //nolint:dupl
	keys, err := o.storage.ExchangeUserKeys().GetKeysByExchangeSlug(ctx, repo_exchange_user_keys.GetKeysByExchangeSlugParams{
		UserID:       userID,
		ExchangeSlug: models.ExchangeSlugGateio,
	})
	if err != nil {
		return nil, err
	}

	keyMap := make(map[string]string)
	for _, key := range keys {
		keyMap[string(key.Name)] = key.Value
	}

	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugGateio)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}
	accessKey := keyMap[models.ExchangeKeyNameAccessKey.String()]
	secretKey := keyMap[models.ExchangeKeyNameSecretKey.String()]

	return gateio.NewService(o.l, accessKey, secretKey, baseURL, o.storage, o.store, o.currConvService, httpClient)
}

func (o *Manager) createGateioServiceRaw(ctx context.Context, accessKey, secretKey string, httpClient *http.Client) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugGateio)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return gateio.NewService(o.l, accessKey, secretKey, baseURL, o.storage, o.store, o.currConvService, httpClient)
}

func (o *Manager) createPublicGateioService(ctx context.Context) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugGateio)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return gateio.NewService(o.l, "-", "-", baseURL, o.storage, o.store, o.currConvService, nil)
}

func (o *Manager) createBingxServiceRaw(ctx context.Context, apiKey, secretKey string, httpClient *http.Client) (IExchangeClient, error) {
	baseURL, err := o.exchangeBaseURL(ctx, models.ExchangeSlugBingx)
	if err != nil {
		return nil, err
	}

	return bingx.NewService(o.l, apiKey, secretKey, baseURL, o.storage, o.store, o.currConvService, httpClient)
}

func (o *Manager) createBingxService(ctx context.Context, userID uuid.UUID, httpClient *http.Client) (IExchangeClient, error) {
	keys, err := o.storage.ExchangeUserKeys().GetKeysByExchangeSlug(ctx, repo_exchange_user_keys.GetKeysByExchangeSlugParams{
		UserID:       userID,
		ExchangeSlug: models.ExchangeSlugBingx,
	})
	if err != nil {
		return nil, err
	}

	keyMap := make(map[string]string, len(keys))
	for _, key := range keys {
		keyMap[string(key.Name)] = key.Value
	}

	baseURL, err := o.exchangeBaseURL(ctx, models.ExchangeSlugBingx)
	if err != nil {
		return nil, err
	}

	return bingx.NewService(
		o.l,
		keyMap[models.ExchangeKeyNameAPIKey.String()],
		keyMap[models.ExchangeKeyNameSecretKey.String()],
		baseURL,
		o.storage,
		o.store,
		o.currConvService,
		httpClient,
	)
}

func (o *Manager) createMexcServiceRaw(ctx context.Context, apiKey, secretKey string, httpClient *http.Client) (IExchangeClient, error) {
	baseURL, err := o.exchangeBaseURL(ctx, models.ExchangeSlugMexc)
	if err != nil {
		return nil, err
	}

	return mexc.NewService(o.l, apiKey, secretKey, baseURL, o.storage, o.store, o.currConvService, httpClient)
}

func (o *Manager) createMexcService(ctx context.Context, userID uuid.UUID, httpClient *http.Client) (IExchangeClient, error) {
	keys, err := o.storage.ExchangeUserKeys().GetKeysByExchangeSlug(ctx, repo_exchange_user_keys.GetKeysByExchangeSlugParams{
		UserID:       userID,
		ExchangeSlug: models.ExchangeSlugMexc,
	})
	if err != nil {
		return nil, err
	}

	keyMap := make(map[string]string, len(keys))
	for _, key := range keys {
		keyMap[string(key.Name)] = key.Value
	}

	baseURL, err := o.exchangeBaseURL(ctx, models.ExchangeSlugMexc)
	if err != nil {
		return nil, err
	}

	return mexc.NewService(
		o.l,
		keyMap[models.ExchangeKeyNameAPIKey.String()],
		keyMap[models.ExchangeKeyNameSecretKey.String()],
		baseURL,
		o.storage,
		o.store,
		o.currConvService,
		httpClient,
	)
}

func (o *Manager) createKrakenServiceRaw(ctx context.Context, apiKey, secretKey string, httpClient *http.Client) (IExchangeClient, error) {
	baseURL, err := o.exchangeBaseURL(ctx, models.ExchangeSlugKraken)
	if err != nil {
		return nil, err
	}

	return kraken.NewService(o.l, apiKey, secretKey, baseURL, o.storage, o.store, o.currConvService, httpClient)
}

func (o *Manager) createKrakenService(ctx context.Context, userID uuid.UUID, httpClient *http.Client) (IExchangeClient, error) {
	keys, err := o.storage.ExchangeUserKeys().GetKeysByExchangeSlug(ctx, repo_exchange_user_keys.GetKeysByExchangeSlugParams{
		UserID:       userID,
		ExchangeSlug: models.ExchangeSlugKraken,
	})
	if err != nil {
		return nil, err
	}

	keyMap := make(map[string]string, len(keys))
	for _, key := range keys {
		keyMap[string(key.Name)] = key.Value
	}

	return o.createKrakenServiceRaw(ctx,
		keyMap[models.ExchangeKeyNameAPIKey.String()],
		keyMap[models.ExchangeKeyNameSecretKey.String()],
		httpClient,
	)
}

// GetNetConfig returns the user's proxy and network interface for the exchange.
// No settings (or no user_exchanges row yet) means a direct connection.
func (o *Manager) GetNetConfig(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) (utils.NetConfig, error) {
	settings, err := o.storage.UserExchanges().GetNetworkSettingsBySlug(ctx, repo_user_exchanges.GetNetworkSettingsBySlugParams{
		UserID: userID,
		Slug:   slug,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return utils.NetConfig{}, nil
	}
	if err != nil {
		return utils.NetConfig{}, fmt.Errorf("get exchange network settings: %w", err)
	}

	return utils.NetConfig{
		ProxyURL:  settings.ProxyUrl.String,
		Interface: settings.NetworkInterface.String,
	}, nil
}

// newHTTPClient returns nil for an empty config, so exchange clients keep their default HTTP client.
func newHTTPClient(netCfg utils.NetConfig) (*http.Client, error) {
	if netCfg == (utils.NetConfig{}) {
		return nil, nil //nolint:nilnil
	}
	return utils.NewHTTPClient(netCfg, exchangeHTTPTimeout)
}

func (o *Manager) exchangeBaseURL(ctx context.Context, slug models.ExchangeSlug) (*url.URL, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return url.Parse(ex.Url)
}
