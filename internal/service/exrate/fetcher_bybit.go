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

type BybitSymbol struct {
	Symbol    string          `json:"symbol"`
	LastPrice decimal.Decimal `json:"lastPrice"`
}

type BybitResponse struct {
	Result struct {
		List []BybitSymbol `json:"list"`
	} `json:"result"`
}

func NewBybitFetcher(url string, proxies []string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &bybitFetcher{url: url, proxies: proxies, httpClient: httpClient, log: log}
}

type bybitFetcher struct {
	url        string
	httpClient *http.Client
	proxies    []string
	log        logger.Logger
}

var _ IFetcher = (*bybitFetcher)(nil)

func (o *bybitFetcher) Source() string {
	return "bybit"
}

func (o *bybitFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error { //nolint:dupl
	err := o.fetchWithClient(ctx, o.httpClient, "direct", currencyFilter, out)
	if err == nil {
		return nil
	}

	o.log.Warnw("[EXRATE-BYBIT] direct request failed, trying proxies", "error", err)

	if len(o.proxies) == 0 {
		o.log.Errorw("[EXRATE-BYBIT] no proxies available after direct failure", "error", err)
		return err
	}

	shuffledProxies := make([]string, len(o.proxies))
	copy(shuffledProxies, o.proxies)
	rand.Shuffle(len(shuffledProxies), func(i, j int) {
		shuffledProxies[i], shuffledProxies[j] = shuffledProxies[j], shuffledProxies[i]
	})

	var lastErr error = err //nolint:all

	for _, proxyURL := range shuffledProxies {
		client, err := o.createProxyClient(proxyURL)
		if err != nil {
			o.log.Warnw("[EXRATE-BYBIT] failed to create proxy client", "proxy", proxyURL, "error", err)
			lastErr = err
			continue
		}

		err = o.fetchWithClient(ctx, client, proxyURL, currencyFilter, out)
		if err == nil {
			o.log.Infow("[EXRATE-BYBIT] request succeeded with proxy", "proxy", proxyURL)
			return nil
		}

		o.log.Warnw("[EXRATE-BYBIT] request failed with proxy, trying next", "proxy", proxyURL, "error", err)
		lastErr = err
	}

	o.log.Errorw("[EXRATE-BYBIT] all proxies exhausted",
		"proxy_count", len(shuffledProxies),
		"last_error", lastErr)
	return fmt.Errorf("all proxies exhausted, last error: %w", lastErr)
}

func (o *bybitFetcher) createProxyClient(proxyURL string) (*http.Client, error) {
	parsedProxy, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(parsedProxy),
	}

	return &http.Client{
		Transport: transport,
		Timeout:   o.httpClient.Timeout,
	}, nil
}

func (o *bybitFetcher) fetchWithClient(ctx context.Context, client *http.Client, connectionType string, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.url, http.NoBody)
	if err != nil {
		o.log.Errorw("[EXRATE-BYBIT] failed to create request",
			"error", err,
			"url", o.url,
			"connection", connectionType)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		o.log.Errorw("[EXRATE-BYBIT] http client error",
			"error", err,
			"url", o.url,
			"connection", connectionType)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		o.log.Errorw("[EXRATE-BYBIT] failed to read response body",
			"error", err,
			"status_code", resp.StatusCode,
			"connection", connectionType)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		o.log.Errorw("[EXRATE-BYBIT] non-OK HTTP status",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"raw_response", string(bodyBytes),
			"connection", connectionType)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := parseBybitResponseBytes(bodyBytes)
	if err != nil {
		o.log.Errorw("[EXRATE-BYBIT] response parsing error",
			"error", err,
			"raw_response", string(bodyBytes),
			"status_code", resp.StatusCode,
			"connection", connectionType)
		return err
	}

	if err := filterBybitResponse(body, currencyFilter, out); err != nil {
		o.log.Errorw("[EXRATE-BYBIT] failed to filter response",
			"error", err,
			"symbol_count", len(body.Result.List),
			"connection", connectionType)
		return err
	}

	o.log.Infow("[EXRATE-BYBIT] successfully fetched exchange rates",
		"symbol_count", len(body.Result.List),
		"connection", connectionType)

	return nil
}

func parseBybitResponseBytes(b []byte) (*BybitResponse, error) {
	r := &BybitResponse{}
	if err := json.Unmarshal(b, r); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}
	return r, nil
}

func filterBybitResponse(r *BybitResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	if r == nil || len(r.Result.List) == 0 {
		return fmt.Errorf("empty response data")
	}

	for _, symbol := range r.Result.List {
		if s, ok := currencyFilter.symbols[symbol.Symbol]; ok {
			if symbol.LastPrice.IsZero() {
				return fmt.Errorf("zero LastPrice for symbol %s", symbol.Symbol)
			}

			if symbol.LastPrice.IsNegative() {
				return fmt.Errorf("negative LastPrice for symbol %s: %s", symbol.Symbol, symbol.LastPrice.String())
			}

			out <- ExRate{
				Source: "bybit",
				From:   s.From,
				To:     s.To,
				Value:  symbol.LastPrice.String(),
			}
			out <- ExRate{
				Source: "bybit",
				From:   s.To,
				To:     s.From,
				Value:  strconv.FormatFloat(1/symbol.LastPrice.InexactFloat64(), 'f', -1, 64),
			}
		}
	}
	return nil
}
