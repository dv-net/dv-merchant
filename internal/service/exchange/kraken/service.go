package kraken

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/tools/hash"
	kraken "github.com/dv-net/dv-merchant/pkg/exchange_client/kraken"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

type Service struct {
	exClient kraken.IKrakenClient
	assets   *assetsCache
	storage  storage.IStorage
	convSvc  currconv.ICurrencyConvertor
	l        logger.Logger
	connHash string
}

func NewService(logger logger.Logger, apiKey, secretKey string, baseURL *url.URL, storage storage.IStorage, store limiter.Store, convSvc currconv.ICurrencyConvertor, httpClient *http.Client) (*Service, error) {
	exClient, err := kraken.NewBaseClient(&kraken.ClientOptions{
		APIKey:    apiKey,
		SecretKey: secretKey,
		BaseURL:   baseURL,
	}, store, kraken.WithLogger(logger), kraken.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	connHash, err := hash.SHA256ConnectionHash(models.ExchangeSlugKraken.String(), apiKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection hash: %w", err)
	}

	return &Service{
		exClient: exClient,
		assets:   assetsCacheFor(baseURL.String(), exClient.Market()),
		storage:  storage,
		convSvc:  convSvc,
		l:        logger,
		connHash: connHash,
	}, nil
}

func (o *Service) TestConnection(ctx context.Context) error {
	if _, err := o.exClient.Account().GetBalance(ctx); err != nil {
		return fmt.Errorf("get balance: %w", err)
	}
	return nil
}

func (o *Service) GetConnectionHash() string {
	return o.connHash
}

func (o *Service) GetAccountBalance(ctx context.Context) ([]*models.AccountBalanceDTO, error) {
	available, err := o.availableBalances(ctx)
	if err != nil {
		return nil, err
	}

	balances := make([]*models.AccountBalanceDTO, 0, len(available))
	for code, amount := range available {
		if !amount.IsPositive() {
			continue
		}

		currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByTicker(ctx, code)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return nil, fmt.Errorf("get internal currency id for %s: %w", code, err)
		}

		amountUSD, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
			Source: models.ExchangeSlugKraken.String(),
			From:   code,
			To:     models.CurrencyCodeUSDT,
			Amount: amount.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("convert %s balance to usdt: %w", code, err)
		}

		balances = append(balances, &models.AccountBalanceDTO{
			Currency:  currencyID,
			Amount:    amount,
			AmountUSD: amountUSD.Round(4),
			Type:      models.CurrencyTypeCrypto.String(),
		})
	}

	return balances, nil
}

func (o *Service) GetCurrencyBalance(ctx context.Context, currency string) (*decimal.Decimal, error) {
	available, err := o.availableBalances(ctx)
	if err != nil {
		return nil, err
	}

	amount, ok := available[currency]
	if !ok {
		return &decimal.Zero, nil
	}
	return &amount, nil
}

// availableBalances returns spendable amounts keyed by our currency code (BTC, USDT, ...).
// Earn balances (USDT.F, XBT.M, ...) are skipped: they can not be traded or withdrawn.
func (o *Service) availableBalances(ctx context.Context) (map[string]decimal.Decimal, error) {
	res, err := o.exClient.Account().GetBalanceEx(ctx)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	assets, err := o.assets.get(ctx)
	if err != nil {
		return nil, err
	}

	balances := make(map[string]decimal.Decimal, len(res))
	for name, balance := range res {
		if isEarnAsset(name) {
			continue
		}
		code := assets.code(name)
		balances[code] = balances[code].Add(balance.Available())
	}
	return balances, nil
}
