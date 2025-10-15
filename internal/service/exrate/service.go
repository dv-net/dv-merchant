package exrate

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/service/currency"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/pkg/key_value"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

type IExRateSource interface {
	Run(ctx context.Context)
	GetCurrencyRate(ctx context.Context, source, from, to string, scale ...decimal.Decimal) (string, error)
	GetAllCurrencyRate(ctx context.Context, source string, scale ...decimal.Decimal) ([]ExRate, error)
	LoadRatesList(ctx context.Context, rateSource string) (*Rates, error)
	LoadSources(ctx context.Context) ([]string, error)
	GetStoreCurrencyRate(ctx context.Context, currencies []*models.Currency, source string, scale ...decimal.Decimal) (map[string]string, error)
}

// New creates new exchange rate source service
func New(
	cfg *config.Config,
	currencyService currency.ICurrency,
	logger logger.Logger,
	storage storage.IStorage,
) (IExRateSource, error) {
	srv := &service{
		storage:         storage,
		currencyService: currencyService,
		fetchInterval:   cfg.Exrate.FetchInterval,
		fetchers:        make(map[string]fetcherData, 8),
		logger:          logger,
	}

	f := NewBinanceFetcher("https://api.binance.com/api/v3/ticker/price", http.DefaultClient, logger)
	srv.fetchers[f.Source()] = fetcherData{
		fetcher: f,
		cfChan:  make(chan CurrencyFilter),
	}

	f = NewHtxFetcher("https://api.huobi.pro/market/tickers", http.DefaultClient, logger)
	srv.fetchers[f.Source()] = fetcherData{
		fetcher: f,
		cfChan:  make(chan CurrencyFilter),
	}

	f = NewOkxFetcher("https://www.okx.com/api/v5/market/index-tickers", http.DefaultClient, logger)
	srv.fetchers[f.Source()] = fetcherData{
		fetcher: f,
		cfChan:  make(chan CurrencyFilter),
	}

	f = NewBitgetFetcher("https://api.bitget.com/api/v2/spot/market/tickers", http.DefaultClient, logger)
	srv.fetchers[f.Source()] = fetcherData{
		fetcher: f,
		cfChan:  make(chan CurrencyFilter),
	}

	f = NewKucoinFetcher("https://api.kucoin.com/api/v1/market/allTickers", http.DefaultClient, logger)
	srv.fetchers[f.Source()] = fetcherData{
		fetcher: f,
		cfChan:  make(chan CurrencyFilter),
	}

	f = NewGateioFetcher("https://api.gateio.ws/api/v4/spot/tickers", http.DefaultClient, logger)
	srv.fetchers[f.Source()] = fetcherData{
		fetcher: f,
		cfChan:  make(chan CurrencyFilter),
	}

	f = NewBybitFetcher("https://api.bybit.com/v5/market/tickers?category=spot", http.DefaultClient, logger)
	srv.fetchers[f.Source()] = fetcherData{
		fetcher: f,
		cfChan:  make(chan CurrencyFilter),
	}

	return srv, nil
}

type service struct {
	storage         storage.IStorage
	currencyService currency.ICurrency
	fetchInterval   time.Duration
	fetchers        map[string]fetcherData
	logger          logger.Logger
}

func (srv *service) LoadSources(_ context.Context) ([]string, error) {
	sources := make([]string, 0, len(srv.fetchers))
	for source := range srv.fetchers {
		sources = append(sources, source)
	}
	return sources, nil
}

var _ IExRateSource = (*service)(nil) // Ensure service support interface

type fetcherData struct {
	fetcher IFetcher
	cfChan  chan CurrencyFilter
}

func (srv *service) Run(ctx context.Context) {
	wg := sync.WaitGroup{}
	outChan := make(chan ExRate)
	wg.Add(1)
	go srv.rateKeeper(ctx, outChan, &wg)
	for _, fd := range srv.fetchers {
		wg.Add(1)
		go srv.fetchWorker(ctx, fd.fetcher, fd.cfChan, outChan, &wg)
	}
	srv.logger.Info("Exchange Rate Service Start")
	timer := time.NewTimer(0)
LOOP:
	for {
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			break LOOP
		case <-timer.C:
			timer.Reset(srv.fetchInterval)
			currencyList, err := srv.getCurrencyList(ctx)
			if err != nil {
				srv.logger.Errorw("get all currency list", "error", err)
				continue LOOP
			}
			cf := NewCurrencyFilter(currencyList)
			for _, fd := range srv.fetchers {
				fd.cfChan <- cf
			}
		}
	}
	close(outChan)
	wg.Wait()
	srv.logger.Info("Exchange Rate Service Done")
}

func (srv *service) rateKeeper(ctx context.Context, in <-chan ExRate, wg *sync.WaitGroup) {
	defer wg.Done()
	ttl := srv.fetchInterval * 10
	for {
		select {
		case <-ctx.Done():
			return
		case rate := <-in:
			if err := srv.storage.CurrencyExchange().StoreRate(ctx, rate.Source, rate.From, rate.To, rate.Value, ttl); err != nil {
				srv.logger.Errorw(
					"save exchange rate ", err,
					"source", rate.Source,
					"from", rate.From,
					"to", rate.To,
					"value", rate.Value,
				)
			} else {
				srv.logger.Debugw(
					"exrate stored",
					"source", rate.Source,
					"from", rate.From,
					"to", rate.To,
					"value", rate.Value,
				)
			}
		}
	}
}

func (srv *service) GetCurrencyRate(ctx context.Context, source, from, to string, scale ...decimal.Decimal) (v string, err error) {
	if !models.RateSource(source).Valid() {
		return "", fmt.Errorf("invalid source %s", source)
	}

	if strings.EqualFold(to, models.CurrencyCodeUSD) {
		to = models.CurrencyCodeUSDT
	}

	if strings.EqualFold(from, to) {
		return "1", nil
	}

	if v, err = srv.storage.CurrencyExchange().GetRate(ctx, source, from, to); err != nil {
		if errors.Is(err, key_value.ErrEntryNotFound) || errors.Is(err, redis.Nil) {
			err = &ExchangeRateNotFoundError{Source: source, From: from, To: to}
		} else {
			err = fmt.Errorf("get currency exchange rate: %w [%s/%s]", err, from, to)
		}
	}
	if v == "" {
		return v, err
	}

	rateValue, parseErr := decimal.NewFromString(v)
	if parseErr != nil {
		srv.logger.Errorw("failed to parse rate", "error", parseErr)
		return v, fmt.Errorf("failed to parse rate: %w", err)
	}

	rateValue = normalizeStableRate(from, to, rateValue)

	valScale := rateValue
	if scale != nil {
		multiplier := decimal.NewFromInt(100).Add(scale[0]).Div(decimal.NewFromInt(100))
		valScale = rateValue.Mul(multiplier)
		return valScale.String(), err
	}
	return v, err
}

func (srv *service) GetAllCurrencyRate(ctx context.Context, source string, scale ...decimal.Decimal) (result []ExRate, err error) {
	if !models.RateSource(source).Valid() {
		return nil, fmt.Errorf("invalid source %s", source)
	}

	currencyList, err := srv.getCurrencyList(ctx)
	if err != nil {
		return nil, err
	}

	filter := NewCurrencyFilter(currencyList)
	result = make([]ExRate, 0, len(filter.symbols))

	for _, pair := range filter.symbols {
		rateStr, err := srv.GetCurrencyRate(ctx, source, pair.From, pair.To)
		if err != nil {
			srv.logger.Debugw("currency rate not exists", "error", err)
			continue
		}

		if rateStr == "" {
			continue
		}

		rateValue, parseErr := decimal.NewFromString(rateStr)
		if parseErr != nil {
			srv.logger.Errorw("failed to parse rate", "error", parseErr)
			continue
		}

		rateValue = normalizeStableRate(pair.From, pair.To, rateValue)

		// Add scale if needed
		valScale := rateValue
		if len(scale) > 0 {
			multiplier := decimal.NewFromInt(100).Add(scale[0]).Div(decimal.NewFromInt(100))
			valScale = rateValue.Mul(multiplier)
		}

		result = append(result, ExRate{
			Source:     source,
			From:       pair.From,
			To:         pair.To,
			UpdatedAt:  util.Pointer(time.Now().UTC()),
			Value:      rateStr,
			ValueScale: valScale.String(),
		})
	}

	slices.SortFunc(result, func(a, b ExRate) int {
		if a.From != b.From {
			if a.From < b.From {
				return -1
			}
			return 1
		}
		if a.To < b.To {
			return -1
		} else if a.To > b.To {
			return 1
		}
		return 0
	})

	return result, nil
}

func (srv *service) LoadRatesList(ctx context.Context, rateSource string) (*Rates, error) {
	if !models.RateSource(rateSource).Valid() {
		return nil, fmt.Errorf("invalid source %s", rateSource)
	}

	currencies, err := srv.currencyService.GetCurrenciesEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled currencies: %w", err)
	}

	var rates Rates
	for _, curr := range currencies {
		if curr.IsStablecoin {
			rates.CurrencyIDs = append(rates.CurrencyIDs, curr.ID)
			rates.Rate = append(rates.Rate, decimal.NewFromInt(1))
			continue
		}
		rate, err := srv.GetCurrencyRate(ctx, rateSource, curr.Code, "USDT")
		if err != nil {
			return nil, fmt.Errorf("failed to get currency rate: %w", err)
		}

		rateNumber, err := decimal.NewFromString(rate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse rate: %w", err)
		}

		rates.CurrencyIDs = append(rates.CurrencyIDs, curr.ID)
		rates.Rate = append(rates.Rate, rateNumber)
	}

	return &rates, nil
}

func (srv *service) getCurrencyList(ctx context.Context) (result CurrencyList, err error) {
	currModelList, err := srv.currencyService.GetCurrenciesEnabled(ctx)
	if err != nil {
		srv.logger.Errorw("get all currency list", "error", err)
		return nil, err
	}

	result = make(CurrencyList, 0, len(currModelList))
	for _, v := range currModelList {
		result = append(result, Currency{
			Code:     v.Code,
			IsStable: v.IsStablecoin,
		})
	}

	slices.SortFunc(result, func(a, b Currency) int {
		if a.Code < b.Code {
			return -1
		}
		if a.Code > b.Code {
			return 1
		}
		return 0
	})

	return slices.CompactFunc(result, func(a, b Currency) bool {
		return a.Code == b.Code
	}), nil
}

func (srv *service) GetStoreCurrencyRate(ctx context.Context, currencies []*models.Currency, source string, scale ...decimal.Decimal) (map[string]string, error) {
	rates := make(map[string]string)
	valScale := decimal.Zero
	if len(scale) > 0 {
		valScale = scale[0]
	}

	for _, cur := range currencies {
		if cur.IsStablecoin {
			rates[cur.ID] = "1"
			continue
		}
		rate, err := srv.GetCurrencyRate(ctx, source, cur.Code, "USDT", valScale)
		if err != nil {
			return nil, fmt.Errorf("failed to get currency rate: %w", err)
		}
		rates[cur.ID] = rate
	}
	return rates, nil
}

func normalizeStableRate(from, to string, rate decimal.Decimal) decimal.Decimal {
	stables := models.CurrencyCodeStableSet()
	_, fromIsStable := stables[strings.ToUpper(from)]
	_, toIsStable := stables[strings.ToUpper(to)]

	if fromIsStable && toIsStable {
		inRange := rate.GreaterThanOrEqual(decimal.NewFromFloat(0.85)) &&
			rate.LessThanOrEqual(decimal.NewFromFloat(1.25))
		if inRange {
			return decimal.NewFromInt(1)
		}
	}
	return rate
}
