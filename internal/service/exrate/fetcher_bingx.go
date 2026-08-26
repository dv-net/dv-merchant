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

type BingxSymbol struct {
	Symbol    string          `json:"symbol"`
	LastPrice decimal.Decimal `json:"lastPrice"`
}

type BingxResponse struct {
	Code int           `json:"code"`
	Msg  string        `json:"msg"`
	Data []BingxSymbol `json:"data"`
}

func NewBingxFetcher(url string, proxies []string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &bingxFetcher{url: url, proxies: proxies, httpClient: httpClient, log: log}
}

type bingxFetcher struct {
	url        string
	httpClient *http.Client
	proxies    []string
	log        logger.Logger
}

var _ IFetcher = (*bingxFetcher)(nil)

func (f *bingxFetcher) Source() string {
	return "bingx"
}

func (f *bingxFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error { //nolint:dupl
	err := f.fetchWithClient(ctx, f.httpClient, "direct", currencyFilter, out)
	if err == nil {
		return nil
	}

	f.log.Warnw("[EXRATE-BINGX] direct request failed, trying proxies", "error", err)

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
			f.log.Warnw("[EXRATE-BINGX] failed to create proxy client", "proxy", proxyURL, "error", err)
			lastErr = err
			continue
		}

		if err := f.fetchWithClient(ctx, client, "proxy", currencyFilter, out); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return lastErr
}

func (f *bingxFetcher) createProxyClient(proxyURL string) (*http.Client, error) {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("parse proxy url: %w", err)
	}

	return &http.Client{
		Timeout:   f.httpClient.Timeout,
		Transport: &http.Transport{Proxy: http.ProxyURL(parsedURL)},
	}, nil
}

func (f *bingxFetcher) fetchWithClient(ctx context.Context, client *http.Client, connectionType string, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.url, http.NoBody)
	if err != nil {
		f.log.Errorw("[EXRATE-BINGX] failed to create request", "error", err, "url", f.url, "connection", connectionType)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		f.log.Errorw("[EXRATE-BINGX] http client error", "error", err, "url", f.url, "connection", connectionType)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		f.log.Errorw("[EXRATE-BINGX] failed to read response body", "error", err, "status_code", resp.StatusCode, "connection", connectionType)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		f.log.Errorw("[EXRATE-BINGX] non-OK HTTP status",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"raw_response", string(bodyBytes),
			"connection", connectionType)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := parseBingxResponseBytes(bodyBytes)
	if err != nil {
		f.log.Errorw("[EXRATE-BINGX] response parsing error", "error", err, "status_code", resp.StatusCode, "connection", connectionType)
		return err
	}

	if err := filterBingxResponse(body, currencyFilter, out); err != nil {
		f.log.Errorw("[EXRATE-BINGX] failed to filter response", "error", err, "symbol_count", len(body.Data), "connection", connectionType)
		return err
	}

	return nil
}

func parseBingxResponseBytes(b []byte) (*BingxResponse, error) {
	r := &BingxResponse{}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}
	if r.Code != 0 {
		return nil, fmt.Errorf("bingx error: %s (%d)", r.Msg, r.Code)
	}
	return r, nil
}

func filterBingxResponse(r *BingxResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	if r == nil || len(r.Data) == 0 {
		return fmt.Errorf("empty response data")
	}

	for _, symbol := range r.Data {
		s, ok := currencyFilter.symbols[removeDashFromSymbol(symbol.Symbol)]
		if !ok {
			continue
		}

		if symbol.LastPrice.IsZero() || symbol.LastPrice.IsNegative() {
			return fmt.Errorf("invalid price for symbol %s: %s", symbol.Symbol, symbol.LastPrice.String())
		}

		out <- ExRate{
			Source: "bingx",
			From:   s.From,
			To:     s.To,
			Value:  symbol.LastPrice.String(),
		}
		out <- ExRate{
			Source: "bingx",
			From:   s.To,
			To:     s.From,
			Value:  strconv.FormatFloat(1/symbol.LastPrice.InexactFloat64(), 'f', -1, 64),
		}
	}
	return nil
}
