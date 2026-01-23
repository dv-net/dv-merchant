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

	"github.com/dv-net/dv-merchant/pkg/logger"
)

var ErrInvalidResponseFromKucoin = errors.New("invalid response from kucoin")

type KucoinSymbol struct {
	Symbol string `json:"symbol"`
	Last   string `json:"last,omitempty"`
}

type KucoinResponse struct {
	Code string `json:"code,omitempty"`
	Data struct {
		Time   int64          `json:"time,omitempty"`
		Ticker []KucoinSymbol `json:"ticker,omitempty"`
	} `json:"data"`
}

func NewKucoinFetcher(url string, proxies []string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &kucoinFetcher{url: url, proxies: proxies, httpClient: httpClient, log: log}
}

type kucoinFetcher struct {
	url        string
	httpClient *http.Client
	proxies    []string
	log        logger.Logger
}

var _ IFetcher = (*kucoinFetcher)(nil)

func (o *kucoinFetcher) Source() string {
	return "kucoin"
}

func (o *kucoinFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error { //nolint:dupl
	err := o.fetchWithClient(ctx, o.httpClient, "direct", currencyFilter, out)
	if err == nil {
		return nil // Success with direct connection
	}

	o.log.Warnw("[EXRATE-KUCOIN] direct request failed, trying proxies", "error", err)

	if len(o.proxies) == 0 {
		o.log.Errorw("[EXRATE-KUCOIN] no proxies available after direct failure", "error", err)
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
			o.log.Warnw("[EXRATE-KUCOIN] failed to create proxy client", "proxy", proxyURL, "error", err)
			lastErr = err
			continue
		}

		err = o.fetchWithClient(ctx, client, proxyURL, currencyFilter, out)
		if err == nil {
			o.log.Infow("[EXRATE-KUCOIN] request succeeded with proxy", "proxy", proxyURL)
			return nil // Success
		}

		o.log.Warnw("[EXRATE-KUCOIN] request failed with proxy, trying next", "proxy", proxyURL, "error", err)
		lastErr = err
	}

	o.log.Errorw("[EXRATE-KUCOIN] all proxies exhausted", "proxy_count", len(shuffledProxies), "last_error", lastErr)
	return fmt.Errorf("all proxies exhausted, last error: %w", lastErr)
}

func (o *kucoinFetcher) createProxyClient(proxyURL string) (*http.Client, error) {
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

func (o *kucoinFetcher) fetchWithClient(ctx context.Context, client *http.Client, connectionType string, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.url, nil)
	if err != nil {
		o.log.Errorw("[EXRATE-KUCOIN] failed to create request",
			"error", err,
			"url", o.url,
			"connection", connectionType)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		o.log.Errorw("[EXRATE-KUCOIN] http client error",
			"error", err,
			"url", o.url,
			"connection", connectionType)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		o.log.Errorw("[EXRATE-KUCOIN] failed to read response body",
			"error", err,
			"status_code", resp.StatusCode,
			"connection", connectionType)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		o.log.Errorw("[EXRATE-KUCOIN] non-OK HTTP status",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"raw_response", string(bodyBytes),
			"connection", connectionType)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := parseKucoinResponseBytes(bodyBytes)
	if err != nil {
		o.log.Errorw("[EXRATE-KUCOIN] response parsing error",
			"error", err,
			"raw_response", string(bodyBytes),
			"status_code", resp.StatusCode,
			"connection", connectionType)
		return err
	}

	if body.Code != "200000" {
		o.log.Errorw(
			"[EXRATE-KUCOIN] currency exchange service response not OK",
			"raw_response", string(bodyBytes),
			"status", body.Code,
			"connection", connectionType,
		)
		return ErrInvalidResponseFromKucoin
	}

	if err := filterKucoinResponse(body, currencyFilter, out); err != nil {
		o.log.Errorw("[EXRATE-KUCOIN] failed to filter response",
			"error", err,
			"ticker_count", len(body.Data.Ticker),
			"connection", connectionType)
		return err
	}

	o.log.Infow("[EXRATE-KUCOIN] successfully fetched exchange rates",
		"ticker_count", len(body.Data.Ticker),
		"connection", connectionType)

	return nil
}

func parseKucoinResponseBytes(b []byte) (*KucoinResponse, error) {
	r := &KucoinResponse{}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}
	return r, nil
}

func filterKucoinResponse(r *KucoinResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	if r == nil || len(r.Data.Ticker) == 0 {
		return fmt.Errorf("empty response data")
	}

	for _, symbol := range r.Data.Ticker {
		if s, ok := currencyFilter.symbols[removeDashFromSymbol(symbol.Symbol)]; ok {
			floatValue, err := strconv.ParseFloat(symbol.Last, 64)
			if err != nil {
				return fmt.Errorf("failed to parse Last for symbol %s: %w", symbol.Symbol, err)
			}

			if floatValue <= 0 {
				return fmt.Errorf("invalid Last price for symbol %s: %f", symbol.Symbol, floatValue)
			}

			out <- ExRate{
				Source: "kucoin",
				From:   s.From,
				To:     s.To,
				Value:  symbol.Last,
			}
			out <- ExRate{
				Source: "kucoin",
				From:   s.To,
				To:     s.From,
				Value:  strconv.FormatFloat(1/floatValue, 'f', -1, 64),
			}
		}
	}
	return nil
}
