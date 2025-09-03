package exchange_manager

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/exchange/binance"
	"github.com/dv-net/dv-merchant/internal/service/exchange/bitget"
	"github.com/dv-net/dv-merchant/internal/service/exchange/bybit"
	"github.com/dv-net/dv-merchant/internal/service/exchange/gateio"
	"github.com/dv-net/dv-merchant/internal/service/exchange/htx"
	"github.com/dv-net/dv-merchant/internal/service/exchange/kucoin"
	"github.com/dv-net/dv-merchant/internal/service/exchange/okx"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_user_keys"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

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
	CreateDriver(ctx context.Context, slug models.ExchangeSlug, apiKey, secretKey, passphrase string) (IExchangeClient, error)
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

func (o *Manager) CreateDriver(ctx context.Context, slug models.ExchangeSlug, apiKey, secretKey, passphrase string) (IExchangeClient, error) {
	if !slug.Valid() {
		return nil, fmt.Errorf("invalid exchange slug provided")
	}
	switch slug {
	case models.ExchangeSlugBinance:
		return o.createBinanceServiceRaw(ctx, apiKey, secretKey)
	case models.ExchangeSlugHtx:
		return o.createHtxServiceRaw(ctx, apiKey, secretKey)
	case models.ExchangeSlugOkx:
		return o.createOkxServiceRaw(ctx, apiKey, secretKey, passphrase)
	case models.ExchangeSlugBitget:
		return o.createBitgetServiceRaw(ctx, apiKey, secretKey, passphrase)
	case models.ExchangeSlugKucoin:
		return o.createKucoinServiceRaw(ctx, apiKey, secretKey, passphrase)
	case models.ExchangeSlugBybit:
		return o.createBybitServiceRaw(ctx, apiKey, secretKey)
	case models.ExchangeSlugGateio:
		return o.createGateioServiceRaw(ctx, apiKey, secretKey)
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

	switch *usr.ExchangeSlug {
	case models.ExchangeSlugHtx:
		return o.createHtxService(ctx, usr.ID)
	case models.ExchangeSlugOkx:
		return o.createOkxService(ctx, userID)
	case models.ExchangeSlugBinance:
		return o.createBinanceService(ctx, userID)
	case models.ExchangeSlugBitget:
		return o.createBitgetService(ctx, userID)
	case models.ExchangeSlugKucoin:
		return o.createKucoinService(ctx, userID)
	case models.ExchangeSlugGateio:
		return o.createGateService(ctx, userID)
	case models.ExchangeSlugBybit:
		return o.createBybitService(ctx, userID)
	default:
		return nil, fmt.Errorf("user is missing current driver")
	}
}

func (o *Manager) GetDriver(ctx context.Context, slug models.ExchangeSlug, userID uuid.UUID) (IExchangeClient, error) {
	if !slug.Valid() {
		return nil, fmt.Errorf("invalid exchange slug provided")
	}
	switch slug {
	case models.ExchangeSlugHtx:
		return o.createHtxService(ctx, userID)
	case models.ExchangeSlugOkx:
		return o.createOkxService(ctx, userID)
	case models.ExchangeSlugBinance:
		return o.createBinanceService(ctx, userID)
	case models.ExchangeSlugBitget:
		return o.createBitgetService(ctx, userID)
	case models.ExchangeSlugKucoin:
		return o.createKucoinService(ctx, userID)
	case models.ExchangeSlugGateio:
		return o.createGateService(ctx, userID)
	case models.ExchangeSlugBybit:
		return o.createBybitService(ctx, userID)
	default:
		return nil, fmt.Errorf("slug %s does not exists", slug.String())
	}
}

func (o *Manager) createOkxServiceRaw(ctx context.Context, apiKey, secretKey, passphrase string) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugOkx)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return okx.NewService(o.l, apiKey, secretKey, passphrase, baseURL, o.storage, o.store, o.currConvService)
}

func (o *Manager) createOkxService(ctx context.Context, userID uuid.UUID) (IExchangeClient, error) {
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

	return okx.NewService(o.l, apiKey, secretKey, passphrase, baseURL, o.storage, o.store, o.currConvService)
}

func (o *Manager) createHtxServiceRaw(ctx context.Context, accessKey, secretKey string) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugHtx)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return htx.NewService(o.l, accessKey, secretKey, baseURL, o.storage, o.store, o.currConvService)
}

func (o *Manager) createHtxService(ctx context.Context, userID uuid.UUID) (IExchangeClient, error) { //nolint:dupl
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

	return htx.NewService(o.l, accessKey, secretKey, baseURL, o.storage, o.store, o.currConvService)
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

	return htx.NewService(o.l, "-", "-", baseURL, o.storage, o.store, o.currConvService)
}

func (o *Manager) createBinanceServiceRaw(ctx context.Context, apiKey, secretKey string) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugBinance)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return binance.NewService(apiKey, secretKey, false, baseURL, o.storage, o.currConvService)
}

func (o *Manager) createBinanceService(ctx context.Context, userID uuid.UUID) (IExchangeClient, error) {
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

	return binance.NewService(accessKey, secretKey, false, baseURL, o.storage, o.currConvService)
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

	return binance.NewService("-", "-", true, baseURL, o.storage, o.currConvService)
}

func (o *Manager) createBitgetServiceRaw(ctx context.Context, accessKey, secretKey, passPhrase string) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugBitget)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return bitget.NewService(o.l, accessKey, secretKey, passPhrase, false, baseURL, o.storage, o.store, o.currConvService)
}

func (o *Manager) createBitgetService(ctx context.Context, userID uuid.UUID) (IExchangeClient, error) { //nolint:dupl
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

	return bitget.NewService(o.l, accessKey, secretKey, passPhrase, false, baseURL, o.storage, o.store, o.currConvService)
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

	return bitget.NewService(o.l, "-", "-", "-", true, baseURL, o.storage, o.store, o.currConvService)
}

func (o *Manager) createKucoinServiceRaw(ctx context.Context, accessKey, secretKey, passPhrase string) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugKucoin)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return kucoin.NewService(o.l, accessKey, secretKey, passPhrase, false, baseURL, o.storage, o.store, o.currConvService)
}

func (o *Manager) createKucoinService(ctx context.Context, userID uuid.UUID) (IExchangeClient, error) { //nolint:dupl
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

	return kucoin.NewService(o.l, accessKey, secretKey, passPhrase, false, baseURL, o.storage, o.store, o.currConvService)
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

	return kucoin.NewService(o.l, "-", "-", "-", true, baseURL, o.storage, o.store, o.currConvService)
}

func (o *Manager) createBybitServiceRaw(ctx context.Context, apiKey, secretKey string) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugBybit)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return bybit.NewService(o.l, apiKey, secretKey, baseURL, o.storage, o.store, o.currConvService)
}

func (o *Manager) createBybitService(ctx context.Context, userID uuid.UUID) (IExchangeClient, error) { //nolint:dupl
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

	return bybit.NewService(o.l, apiKey, secretKey, baseURL, o.storage, o.store, o.currConvService)
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

	return bybit.NewService(o.l, "-", "-", baseURL, o.storage, o.store, o.currConvService)
}

func (o *Manager) createGateService(ctx context.Context, userID uuid.UUID) (IExchangeClient, error) { //nolint:dupl
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

	return gateio.NewService(o.l, accessKey, secretKey, baseURL, o.storage, o.store, o.currConvService)
}

func (o *Manager) createGateioServiceRaw(ctx context.Context, accessKey, secretKey string) (IExchangeClient, error) {
	ex, err := o.storage.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlugGateio)
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(ex.Url)
	if err != nil {
		return nil, err
	}

	return gateio.NewService(o.l, accessKey, secretKey, baseURL, o.storage, o.store, o.currConvService)
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

	return gateio.NewService(o.l, "-", "-", baseURL, o.storage, o.store, o.currConvService)
}
