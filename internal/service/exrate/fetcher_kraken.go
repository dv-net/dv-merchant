package exrate

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	kraken "github.com/dv-net/dv-merchant/pkg/exchange_client/kraken"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/responses"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

const (
	krakenSource        = "kraken"
	krakenPairsCacheTTL = time.Hour
)

func NewKrakenFetcher(baseURL string, proxies []string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &krakenFetcher{baseURL: baseURL, proxies: proxies, httpClient: httpClient, log: log}
}

// krakenFetcher takes last trade prices from /0/public/Ticker. Ticker is keyed by Kraken
// pair name (XXBTZUSD, XBTUSDT), so pair currencies are resolved via /0/public/AssetPairs.
type krakenFetcher struct {
	baseURL    string
	httpClient *http.Client
	proxies    []string
	log        logger.Logger

	mu      sync.Mutex
	pairs   map[string]CurrencyPair // Kraken pair name -> our currency codes
	pairsAt time.Time
}

var _ IFetcher = (*krakenFetcher)(nil)

func (f *krakenFetcher) Source() string {
	return krakenSource
}

// Fetch tries the direct connection first, then configured proxies in random order.
func (f *krakenFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	proxies := make([]string, len(f.proxies))
	copy(proxies, f.proxies)
	rand.Shuffle(len(proxies), func(i, j int) { proxies[i], proxies[j] = proxies[j], proxies[i] })

	var lastErr error
	for i := -1; i < len(proxies); i++ {
		client, connectionType := f.httpClient, "direct"
		if i >= 0 {
			proxyURL, err := url.Parse(proxies[i])
			if err != nil {
				f.log.Debugw("[EXRATE-KRAKEN] invalid proxy url", "proxy", proxies[i], "error", err)
				lastErr = err
				continue
			}
			client = &http.Client{Timeout: f.httpClient.Timeout, Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}
			connectionType = "proxy"
		}

		err := f.fetchWithClient(ctx, client, connectionType, currencyFilter, out)
		if err == nil {
			return nil
		}
		if i < 0 && len(proxies) > 0 {
			f.log.Debugw("[EXRATE-KRAKEN] direct request failed, trying proxies", "error", err)
		}
		lastErr = err
	}

	return lastErr
}

func (f *krakenFetcher) fetchWithClient(ctx context.Context, client *http.Client, connectionType string, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	market, err := f.marketClient(client)
	if err != nil {
		f.log.Warnw("[EXRATE-KRAKEN] failed to create client", "error", err, "connection", connectionType)
		return err
	}

	pairs, err := f.pairCurrencies(ctx, market)
	if err != nil {
		f.log.Warnw("[EXRATE-KRAKEN] failed to load asset pairs", "error", err, "connection", connectionType)
		return err
	}

	tickers, err := market.GetTicker(ctx)
	if err != nil {
		f.log.Warnw("[EXRATE-KRAKEN] failed to load tickers", "error", err, "connection", connectionType)
		return err
	}

	if err := f.filterKrakenResponse(tickers, pairs, currencyFilter, out); err != nil {
		f.log.Warnw("[EXRATE-KRAKEN] failed to filter response", "error", err, "symbol_count", len(tickers), "connection", connectionType)
		return err
	}

	return nil
}

func (f *krakenFetcher) marketClient(httpClient *http.Client) (kraken.IKrakenMarket, error) {
	baseURL, err := url.Parse(f.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	c, err := kraken.NewClient(&kraken.ClientOptions{BaseURL: baseURL}, nil, kraken.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	return kraken.NewMarketClient(c), nil
}

func (f *krakenFetcher) pairCurrencies(ctx context.Context, market kraken.IKrakenMarket) (map[string]CurrencyPair, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.pairs != nil && time.Since(f.pairsAt) < krakenPairsCacheTTL {
		return f.pairs, nil
	}

	res, err := market.GetAssetPairs(ctx)
	if err != nil {
		if f.pairs != nil {
			return f.pairs, nil
		}
		return nil, err
	}

	f.pairs, f.pairsAt = parseKrakenPairs(res), time.Now()
	return f.pairs, nil
}

// parseKrakenPairs maps pair names to our currency codes using wsname ("XBT/USDT" -> BTC, USDT).
func parseKrakenPairs(res responses.AssetPairsResponse) map[string]CurrencyPair {
	pairs := make(map[string]CurrencyPair, len(res))
	for name, pair := range res {
		base, quote, ok := strings.Cut(pair.WsName, "/")
		if !ok || base == "" || quote == "" {
			continue
		}
		pairs[name] = CurrencyPair{From: kraken.CurrencyCode(base), To: kraken.CurrencyCode(quote)}
	}
	return pairs
}

func (f *krakenFetcher) filterKrakenResponse(tickers responses.TickerResponse, pairs map[string]CurrencyPair, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	if len(tickers) == 0 {
		return fmt.Errorf("empty response data")
	}

	for name, ticker := range tickers {
		pair, ok := pairs[name]
		if !ok {
			continue
		}
		s, ok := currencyFilter.FindSymbol(pair.From + pair.To)
		if !ok {
			continue
		}

		price, err := ticker.LastPrice()
		if err != nil || !price.IsPositive() {
			f.log.Debugw("[EXRATE-KRAKEN] skipping pair with invalid price", "pair", name, "price", price.String())
			continue
		}

		out <- ExRate{
			Source: krakenSource,
			From:   s.From,
			To:     s.To,
			Value:  price.String(),
		}
		out <- ExRate{
			Source: krakenSource,
			From:   s.To,
			To:     s.From,
			Value:  decimal.NewFromInt(1).Div(price).String(),
		}
	}
	return nil
}
