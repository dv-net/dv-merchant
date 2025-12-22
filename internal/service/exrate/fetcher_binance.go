//nolint:tagliatelle
package exrate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/shopspring/decimal"
)

type BinanceSymbol struct {
	Symbol string          `json:"symbol"`
	Price  decimal.Decimal `json:"price"`
}

type BinanceResponse []BinanceSymbol

func NewBinanceFetcher(url string, proxies []string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &binanceFetcher{url: url, proxies: proxies, httpClient: httpClient, log: log}
}

type binanceFetcher struct {
	url        string
	httpClient *http.Client
	proxies    []string
	log        logger.Logger
}

var _ IFetcher = (*binanceFetcher)(nil)

func (f *binanceFetcher) Source() string {
	return "binance"
}

func (f *binanceFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	err := f.fetchWithClient(ctx, f.httpClient, currencyFilter, out)
	if err == nil {
		return nil // Success with direct connection
	}

	f.log.Debugw("direct request failed, trying proxies", "error", err)

	if len(f.proxies) == 0 {
		return err
	}

	shuffledProxies := make([]string, len(f.proxies))
	copy(shuffledProxies, f.proxies)
	rand.Shuffle(len(shuffledProxies), func(i, j int) {
		shuffledProxies[i], shuffledProxies[j] = shuffledProxies[j], shuffledProxies[i]
	})

	var lastErr error = err

	for _, proxyURL := range shuffledProxies {
		client, err := f.createProxyClient(proxyURL)
		if err != nil {
			f.log.Debugw("failed to create proxy client", "proxy", proxyURL, "error", err)
			lastErr = err
			continue
		}

		err = f.fetchWithClient(ctx, client, currencyFilter, out)
		if err == nil {
			f.log.Debugw("request succeeded with proxy", "proxy", proxyURL)
			return nil // Success
		}

		f.log.Debugw("request failed with proxy, trying next", "proxy", proxyURL, "error", err)
		lastErr = err
	}

	return fmt.Errorf("all proxies exhausted, last error: %w", lastErr)
}

func (f *binanceFetcher) createProxyClient(proxyURL string) (*http.Client, error) {
	parsedProxy, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(parsedProxy),
	}

	return &http.Client{
		Transport: transport,
		Timeout:   f.httpClient.Timeout,
	}, nil
}

func (f *binanceFetcher) fetchWithClient(ctx context.Context, client *http.Client, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.url, http.NoBody)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := parseBinanceResponse(resp.Body)
	if err != nil {
		return err
	}

	return filterBinanceResponse(body, currencyFilter, out)
}

func parseBinanceResponse(rc io.ReadCloser) (*BinanceResponse, error) {
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	r := &BinanceResponse{}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	return r, nil
}

func filterBinanceResponse(r *BinanceResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error { //nolint:unparam
	for _, symbol := range *r {
		if s, ok := currencyFilter.symbols[symbol.Symbol]; ok {
			out <- ExRate{
				Source: "binance",
				From:   s.From,
				To:     s.To,
				Value:  symbol.Price.String(),
			}
			out <- ExRate{
				Source: "binance",
				From:   s.To,
				To:     s.From,
				Value:  strconv.FormatFloat(1/symbol.Price.InexactFloat64(), 'f', -1, 64),
			}
		}
	}
	return nil
}
