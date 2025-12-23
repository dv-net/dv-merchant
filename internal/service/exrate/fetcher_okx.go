//nolint:tagliatelle
package exrate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/dv-net/dv-merchant/pkg/logger"
)

var ErrInvalidResponseFromOkx = errors.New("invalid response from okx")

type OkxSymbol struct {
	InstID  string `json:"instId"`
	IdxPx   string `json:"idxPx"`
	High24H string `json:"high24h"`
	SodUtc0 string `json:"sodUtc0"`
	Open24H string `json:"open24h"`
	Low24H  string `json:"low24h"`
	SodUtc8 string `json:"sodUtc8"`
	TS      string `json:"ts"`
}

type OkxResponse struct {
	Code    string      `json:"code,omitempty"`
	Message string      `json:"msg,omitempty"`
	Data    []OkxSymbol `json:"data,omitempty"`
}

func NewOkxFetcher(url string, proxies []string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &okxFetcher{url: url, proxies: proxies, httpClient: httpClient, log: log}
}

type okxFetcher struct {
	url        string
	httpClient *http.Client
	proxies    []string
	log        logger.Logger
}

func (o *okxFetcher) Source() string {
	return "okx"
}

func (o *okxFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	err := o.fetchAllCurrencies(ctx, o.httpClient, "direct", currencyFilter, out)
	if err == nil {
		return nil
	}

	o.log.Warnw("[EXRATE-OKX] direct request failed, trying proxies", "error", err)

	if len(o.proxies) == 0 {
		o.log.Errorw("[EXRATE-OKX] no proxies available after direct failure", "error", err)
		return err
	}

	shuffledProxies := make([]string, len(o.proxies))
	copy(shuffledProxies, o.proxies)
	rand.Shuffle(len(shuffledProxies), func(i, j int) {
		shuffledProxies[i], shuffledProxies[j] = shuffledProxies[j], shuffledProxies[i]
	})

	var lastErr error = err

	for _, proxyURL := range shuffledProxies {
		client, err := o.createProxyClient(proxyURL)
		if err != nil {
			o.log.Warnw("[EXRATE-OKX] failed to create proxy client", "proxy", proxyURL, "error", err)
			lastErr = err
			continue
		}

		err = o.fetchAllCurrencies(ctx, client, proxyURL, currencyFilter, out)
		if err == nil {
			o.log.Infow("[EXRATE-OKX] request succeeded with proxy", "proxy", proxyURL)
			return nil
		}

		o.log.Warnw("[EXRATE-OKX] request failed with proxy, trying next", "proxy", proxyURL, "error", err)
		lastErr = err
	}

	o.log.Errorw("[EXRATE-OKX] all proxies exhausted",
		"proxy_count", len(shuffledProxies),
		"last_error", lastErr)
	return fmt.Errorf("all proxies exhausted, last error: %w", lastErr)
}

func (o *okxFetcher) createProxyClient(proxyURL string) (*http.Client, error) {
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

func (o *okxFetcher) fetchAllCurrencies(ctx context.Context, client *http.Client, connectionType string, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	mergeCh := make(chan ExRate, 100)
	errCh := make(chan error, 3)

	ccys := []string{"USDT", "USDC", "BTC"}
	var wg sync.WaitGroup
	wg.Add(len(ccys))

	for _, currency := range ccys {
		go func(currency string) {
			defer wg.Done()
			if err := o.fetchForCurrency(ctx, client, connectionType, currency, currencyFilter, mergeCh); err != nil {
				o.log.Errorw("[EXRATE-OKX] failed to fetch for currency",
					"error", err,
					"currency", currency,
					"connection", connectionType)
				errCh <- fmt.Errorf("currency %s: %w", currency, err)
			} else {
				o.log.Infow("[EXRATE-OKX] successfully fetched for currency",
					"currency", currency,
					"connection", connectionType)
			}
		}(currency)
	}

	go func() {
		wg.Wait()
		close(mergeCh)
		close(errCh)
	}()

	for rate := range mergeCh {
		out <- rate
	}

	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) == len(ccys) {
		o.log.Errorw("[EXRATE-OKX] all currency fetches failed",
			"error_count", len(errs),
			"errors", errs,
			"connection", connectionType)
		return fmt.Errorf("all OKX fetches failed: %v", errs)
	}

	if len(errs) > 0 {
		o.log.Warnw("[EXRATE-OKX] partial fetch failures",
			"failed_count", len(errs),
			"success_count", len(ccys)-len(errs),
			"errors", errs,
			"connection", connectionType)
	}

	return nil
}

func (o *okxFetcher) fetchForCurrency(ctx context.Context, client *http.Client, connectionType string, currency string, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("quoteCcy", currency)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http client error: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	body, err := parseOkxResponseBytes(bodyBytes)
	if err != nil {
		return fmt.Errorf("response parsing error: %w, raw: %s", err, string(bodyBytes))
	}

	if body.Code != "0" {
		return fmt.Errorf("%w: code=%s, msg=%s", ErrInvalidResponseFromOkx, body.Code, body.Message)
	}

	if err := filterOkxResponse(body, currencyFilter, out); err != nil {
		return fmt.Errorf("failed to filter response: %w", err)
	}

	return nil
}

func parseOkxResponseBytes(b []byte) (*OkxResponse, error) {
	r := &OkxResponse{}
	if err := json.Unmarshal(b, r); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}
	return r, nil
}

func filterOkxResponse(r *OkxResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	if r == nil || len(r.Data) == 0 {
		return fmt.Errorf("empty response data")
	}

	for _, symbol := range r.Data {
		if s, ok := currencyFilter.symbols[removeDashFromSymbol(symbol.InstID)]; ok {
			floatValue, err := strconv.ParseFloat(symbol.IdxPx, 64)
			if err != nil {
				return fmt.Errorf("failed to parse IdxPx for symbol %s: %w", symbol.InstID, err)
			}

			if floatValue <= 0 {
				return fmt.Errorf("invalid IdxPx for symbol %s: %f", symbol.InstID, floatValue)
			}

			out <- ExRate{
				Source: "okx",
				From:   s.From,
				To:     s.To,
				Value:  symbol.IdxPx,
			}
			out <- ExRate{
				Source: "okx",
				From:   s.To,
				To:     s.From,
				Value:  strconv.FormatFloat(1/floatValue, 'f', -1, 64),
			}
		}
	}

	return nil
}

func removeDashFromSymbol(symbol string) string {
	return strings.ReplaceAll(symbol, "-", "")
}
