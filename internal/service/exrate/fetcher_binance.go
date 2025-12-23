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
	err := f.fetchWithClient(ctx, f.httpClient, "direct", currencyFilter, out)
	if err == nil {
		return nil // Success with direct connection
	}

	f.log.Warnw("[EXRATE-BINANCE] direct request failed, trying proxies", "error", err)

	if len(f.proxies) == 0 {
		f.log.Errorw("[EXRATE-BINANCE] no proxies available after direct failure", "error", err)
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
			f.log.Warnw("[EXRATE-BINANCE] failed to create proxy client", "proxy", proxyURL, "error", err)
			lastErr = err
			continue
		}

		err = f.fetchWithClient(ctx, client, proxyURL, currencyFilter, out)
		if err == nil {
			f.log.Infow("[EXRATE-BINANCE] request succeeded with proxy", "proxy", proxyURL)
			return nil // Success
		}

		f.log.Warnw("[EXRATE-BINANCE] request failed with proxy, trying next", "proxy", proxyURL, "error", err)
		lastErr = err
	}

	f.log.Errorw("[EXRATE-BINANCE] all proxies exhausted",
		"proxy_count", len(shuffledProxies),
		"last_error", lastErr)
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

func (f *binanceFetcher) fetchWithClient(ctx context.Context, client *http.Client, connectionType string, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.url, http.NoBody)
	if err != nil {
		f.log.Errorw("[EXRATE-BINANCE] failed to create request",
			"error", err,
			"url", f.url,
			"connection", connectionType)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		f.log.Errorw("[EXRATE-BINANCE] http client error",
			"error", err,
			"url", f.url,
			"connection", connectionType)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		f.log.Errorw("[EXRATE-BINANCE] failed to read response body",
			"error", err,
			"status_code", resp.StatusCode,
			"connection", connectionType)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		f.log.Errorw("[EXRATE-BINANCE] non-OK HTTP status",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"raw_response", string(bodyBytes),
			"connection", connectionType)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := parseBinanceResponseBytes(bodyBytes)
	if err != nil {
		f.log.Errorw("[EXRATE-BINANCE] response parsing error",
			"error", err,
			"raw_response", string(bodyBytes),
			"status_code", resp.StatusCode,
			"connection", connectionType)
		return err
	}

	if err := filterBinanceResponse(body, currencyFilter, out); err != nil {
		f.log.Errorw("[EXRATE-BINANCE] failed to filter response",
			"error", err,
			"symbol_count", len(*body),
			"connection", connectionType)
		return err
	}

	f.log.Infow("[EXRATE-BINANCE] successfully fetched exchange rates",
		"symbol_count", len(*body),
		"connection", connectionType)

	return nil
}

func parseBinanceResponseBytes(b []byte) (*BinanceResponse, error) {
	r := &BinanceResponse{}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}
	return r, nil
}

func filterBinanceResponse(r *BinanceResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	if r == nil || len(*r) == 0 {
		return fmt.Errorf("empty response data")
	}

	for _, symbol := range *r {
		if s, ok := currencyFilter.symbols[symbol.Symbol]; ok {
			// Validate price
			if symbol.Price.IsZero() {
				return fmt.Errorf("zero price for symbol %s", symbol.Symbol)
			}

			if symbol.Price.IsNegative() {
				return fmt.Errorf("negative price for symbol %s: %s", symbol.Symbol, symbol.Price.String())
			}

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