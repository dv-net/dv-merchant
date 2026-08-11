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

type MexcSymbol struct {
	Symbol string          `json:"symbol"`
	Price  decimal.Decimal `json:"price"`
}

type MexcResponse []MexcSymbol

func NewMexcFetcher(url string, proxies []string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &mexcFetcher{url: url, proxies: proxies, httpClient: httpClient, log: log}
}

type mexcFetcher struct {
	url        string
	httpClient *http.Client
	proxies    []string
	log        logger.Logger
}

var _ IFetcher = (*mexcFetcher)(nil)

func (f *mexcFetcher) Source() string {
	return "mexc"
}

func (f *mexcFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error { //nolint:dupl
	err := f.fetchWithClient(ctx, f.httpClient, "direct", currencyFilter, out)
	if err == nil {
		return nil
	}

	f.log.Warnw("[EXRATE-MEXC] direct request failed, trying proxies", "error", err)

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
			f.log.Warnw("[EXRATE-MEXC] failed to create proxy client", "proxy", proxyURL, "error", err)
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

func (f *mexcFetcher) createProxyClient(proxyURL string) (*http.Client, error) {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("parse proxy url: %w", err)
	}

	return &http.Client{
		Timeout:   f.httpClient.Timeout,
		Transport: &http.Transport{Proxy: http.ProxyURL(parsedURL)},
	}, nil
}

func (f *mexcFetcher) fetchWithClient(ctx context.Context, client *http.Client, connectionType string, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.url, http.NoBody)
	if err != nil {
		f.log.Errorw("[EXRATE-MEXC] failed to create request", "error", err, "url", f.url, "connection", connectionType)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		f.log.Errorw("[EXRATE-MEXC] http client error", "error", err, "url", f.url, "connection", connectionType)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		f.log.Errorw("[EXRATE-MEXC] failed to read response body", "error", err, "status_code", resp.StatusCode, "connection", connectionType)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		f.log.Errorw("[EXRATE-MEXC] non-OK HTTP status",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"raw_response", string(bodyBytes),
			"connection", connectionType)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := parseMexcResponseBytes(bodyBytes)
	if err != nil {
		f.log.Errorw("[EXRATE-MEXC] response parsing error", "error", err, "status_code", resp.StatusCode, "connection", connectionType)
		return err
	}

	if err := filterMexcResponse(body, currencyFilter, out); err != nil {
		f.log.Errorw("[EXRATE-MEXC] failed to filter response", "error", err, "symbol_count", len(*body), "connection", connectionType)
		return err
	}

	return nil
}

func parseMexcResponseBytes(b []byte) (*MexcResponse, error) {
	r := &MexcResponse{}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}
	return r, nil
}

func filterMexcResponse(r *MexcResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	if r == nil || len(*r) == 0 {
		return fmt.Errorf("empty response data")
	}

	for _, symbol := range *r {
		s, ok := currencyFilter.symbols[symbol.Symbol]
		if !ok {
			continue
		}

		if symbol.Price.IsZero() || symbol.Price.IsNegative() {
			return fmt.Errorf("invalid price for symbol %s: %s", symbol.Symbol, symbol.Price.String())
		}

		out <- ExRate{
			Source: "mexc",
			From:   s.From,
			To:     s.To,
			Value:  symbol.Price.String(),
		}
		out <- ExRate{
			Source: "mexc",
			From:   s.To,
			To:     s.From,
			Value:  strconv.FormatFloat(1/symbol.Price.InexactFloat64(), 'f', -1, 64),
		}
	}
	return nil
}
