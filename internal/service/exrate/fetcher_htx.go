//nolint:tagliatelle
package exrate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dv-net/dv-merchant/pkg/logger"
)

var ErrInvalidResponseFromHtx = errors.New("invalid response from htx")

type HtxSymbol struct {
	Symbol  string  `json:"symbol"`
	Open    float64 `json:"open"`
	High    float64 `json:"high"`
	Low     float64 `json:"low"`
	Close   float64 `json:"close"`
	Amount  float64 `json:"amount"`
	Vol     float64 `json:"vol"`
	Count   int     `json:"count"`
	Bid     float64 `json:"bid"`
	BidSize float64 `json:"bidSize"`
	Ask     float64 `json:"ask"`
	AskSize float64 `json:"askSize"`
}

type HtxResponse struct {
	Status  string      `json:"status,omitempty"`
	Data    []HtxSymbol `json:"data,omitempty"`
	ErrCode string      `json:"err-code,omitempty"`
	ErrMsg  string      `json:"err-msg,omitempty"`
}

func NewHtxFetcher(url string, proxies []string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &htxFetcher{url: url, proxies: proxies, httpClient: httpClient, log: log}
}

type htxFetcher struct {
	url        string
	httpClient *http.Client
	proxies    []string
	log        logger.Logger
}

func (o *htxFetcher) Source() string {
	return "htx"
}

func (o *htxFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error { //nolint:dupl
	err := o.fetchWithClient(ctx, o.httpClient, "direct", currencyFilter, out)
	if err == nil {
		return nil
	}

	o.log.Warnw("[EXRATE-HTX] direct request failed, trying proxies", "error", err)

	if len(o.proxies) == 0 {
		o.log.Errorw("[EXRATE-HTX] no proxies available after direct failure", "error", err)
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
			o.log.Warnw("[EXRATE-HTX] failed to create proxy client", "proxy", proxyURL, "error", err)
			lastErr = err
			continue
		}

		err = o.fetchWithClient(ctx, client, proxyURL, currencyFilter, out)
		if err == nil {
			o.log.Infow("[EXRATE-HTX] request succeeded with proxy", "proxy", proxyURL)
			return nil
		}

		o.log.Warnw("[EXRATE-HTX] request failed with proxy, trying next", "proxy", proxyURL, "error", err)
		lastErr = err
	}

	o.log.Errorw("[EXRATE-HTX] all proxies exhausted",
		"proxy_count", len(shuffledProxies),
		"last_error", lastErr)
	return fmt.Errorf("all proxies exhausted, last error: %w", lastErr)
}

func (o *htxFetcher) createProxyClient(proxyURL string) (*http.Client, error) {
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

func (o *htxFetcher) fetchWithClient(ctx context.Context, client *http.Client, connectionType string, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.url, nil)
	if err != nil {
		o.log.Errorw("[EXRATE-HTX] failed to create request",
			"error", err,
			"url", o.url,
			"connection", connectionType)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		o.log.Errorw("[EXRATE-HTX] http client error",
			"error", err,
			"url", o.url,
			"connection", connectionType)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		o.log.Errorw("[EXRATE-HTX] failed to read response body",
			"error", err,
			"status_code", resp.StatusCode,
			"connection", connectionType)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		o.log.Errorw("[EXRATE-HTX] non-OK HTTP status",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"raw_response", string(bodyBytes),
			"connection", connectionType)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := parseHtxResponseBytes(bodyBytes)
	if err != nil {
		o.log.Errorw("[EXRATE-HTX] response parsing error",
			"error", err,
			"raw_response", string(bodyBytes),
			"status_code", resp.StatusCode,
			"connection", connectionType)
		return err
	}

	if body.Status != "ok" {
		o.log.Errorw(
			"[EXRATE-HTX] currency exchange service response not OK",
			"status", body.Status,
			"err_code", body.ErrCode,
			"err_msg", body.ErrMsg,
			"raw_response", string(bodyBytes),
			"connection", connectionType)
		return fmt.Errorf("%w: %s - %s", ErrInvalidResponseFromHtx, body.ErrCode, body.ErrMsg)
	}

	if err := filterHtxResponse(body, currencyFilter, out); err != nil {
		o.log.Errorw("[EXRATE-HTX] failed to filter response",
			"error", err,
			"symbol_count", len(body.Data),
			"connection", connectionType)
		return err
	}

	o.log.Infow("[EXRATE-HTX] successfully fetched exchange rates",
		"symbol_count", len(body.Data),
		"connection", connectionType)

	return nil
}

func parseHtxResponseBytes(b []byte) (*HtxResponse, error) {
	r := &HtxResponse{}
	if err := json.Unmarshal(b, r); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}
	return r, nil
}

func filterHtxResponse(r *HtxResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	if r == nil || len(r.Data) == 0 {
		return fmt.Errorf("empty response data")
	}

	for _, symbol := range r.Data {
		if s, ok := currencyFilter.symbols[strings.ToUpper(symbol.Symbol)]; ok {
			if symbol.Ask <= 0 {
				return fmt.Errorf("invalid Ask price for symbol %s: %f", symbol.Symbol, symbol.Ask)
			}

			if math.IsNaN(symbol.Ask) || math.IsInf(symbol.Ask, 0) {
				return fmt.Errorf("invalid Ask price (NaN/Inf) for symbol %s", symbol.Symbol)
			}

			askRounded := strconv.FormatFloat(roundToSixDecimalPlaces(symbol.Ask), 'f', 6, 64)
			out <- ExRate{
				Source: "htx",
				From:   s.From,
				To:     s.To,
				Value:  askRounded,
			}
			out <- ExRate{
				Source: "htx",
				From:   s.To,
				To:     s.From,
				Value:  strconv.FormatFloat(roundToSixDecimalPlaces(1/symbol.Ask), 'f', 6, 64),
			}
		}
	}
	return nil
}

func roundToSixDecimalPlaces(value float64) float64 {
	return math.Round(value*1e6) / 1e6
}
