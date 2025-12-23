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
	"strings"

	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/shopspring/decimal"
)

type GateioSymbol struct {
	CurrencyPair string          `json:"currency_pair"`
	Last         decimal.Decimal `json:"last"`
}

type GateioResponse struct {
	Data []GateioSymbol `json:"data"`
}

func NewGateioFetcher(url string, proxies []string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &gateioFetcher{url: url, proxies: proxies, httpClient: httpClient, log: log}
}

type gateioFetcher struct {
	url        string
	httpClient *http.Client
	proxies    []string
	log        logger.Logger
}

var _ IFetcher = (*gateioFetcher)(nil)

func (o *gateioFetcher) Source() string {
	return "gate"
}

func (o *gateioFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	err := o.fetchWithClient(ctx, o.httpClient, "direct", currencyFilter, out)
	if err == nil {
		return nil
	}

	o.log.Warnw("[EXRATE-GATE] direct request failed, trying proxies", "error", err)

	if len(o.proxies) == 0 {
		o.log.Errorw("[EXRATE-GATE] no proxies available after direct failure", "error", err)
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
			o.log.Warnw("[EXRATE-GATE] failed to create proxy client", "proxy", proxyURL, "error", err)
			lastErr = err
			continue
		}

		err = o.fetchWithClient(ctx, client, proxyURL, currencyFilter, out)
		if err == nil {
			o.log.Infow("[EXRATE-GATE] request succeeded with proxy", "proxy", proxyURL)
			return nil
		}

		o.log.Warnw("[EXRATE-GATE] request failed with proxy, trying next", "proxy", proxyURL, "error", err)
		lastErr = err
	}

	o.log.Errorw("[EXRATE-GATE] all proxies exhausted",
		"proxy_count", len(shuffledProxies),
		"last_error", lastErr)
	return fmt.Errorf("all proxies exhausted, last error: %w", lastErr)
}

func (o *gateioFetcher) createProxyClient(proxyURL string) (*http.Client, error) {
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

func (o *gateioFetcher) fetchWithClient(ctx context.Context, client *http.Client, connectionType string, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.url, nil)
	if err != nil {
		o.log.Errorw("[EXRATE-GATE] failed to create request",
			"error", err,
			"url", o.url,
			"connection", connectionType)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		o.log.Errorw("[EXRATE-GATE] http client error",
			"error", err,
			"url", o.url,
			"connection", connectionType)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		o.log.Errorw("[EXRATE-GATE] failed to read response body",
			"error", err,
			"status_code", resp.StatusCode,
			"connection", connectionType)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		o.log.Errorw("[EXRATE-GATE] non-OK HTTP status",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"raw_response", string(bodyBytes),
			"connection", connectionType)
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := parseGateioResponseBytes(bodyBytes)
	if err != nil {
		o.log.Errorw("[EXRATE-GATE] response parsing error",
			"error", err,
			"raw_response", string(bodyBytes),
			"status_code", resp.StatusCode,
			"connection", connectionType)
		return err
	}

	if err := filterGateioResponse(body, currencyFilter, out); err != nil {
		o.log.Errorw("[EXRATE-GATE] failed to filter response",
			"error", err,
			"symbol_count", len(body.Data),
			"connection", connectionType)
		return err
	}

	o.log.Infow("[EXRATE-GATE] successfully fetched exchange rates",
		"symbol_count", len(body.Data),
		"connection", connectionType)

	return nil
}

func parseGateioResponseBytes(b []byte) (*GateioResponse, error) {
	r := &GateioResponse{}
	if err := json.Unmarshal(b, &r.Data); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}
	return r, nil
}

func filterGateioResponse(r *GateioResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	if r == nil || len(r.Data) == 0 {
		return fmt.Errorf("empty response data")
	}

	for _, symbol := range r.Data {
		if s, ok := currencyFilter.symbols[removeGateioDashFromSymbol(symbol.CurrencyPair)]; ok {
			if symbol.Last.IsZero() {
				return fmt.Errorf("zero Last price for currency pair %s", symbol.CurrencyPair)
			}

			if symbol.Last.IsNegative() {
				return fmt.Errorf("negative Last price for currency pair %s: %s",
					symbol.CurrencyPair, symbol.Last.String())
			}

			out <- ExRate{
				Source: "gate",
				From:   s.From,
				To:     s.To,
				Value:  symbol.Last.String(),
			}
			out <- ExRate{
				Source: "gate",
				From:   s.To,
				To:     s.From,
				Value:  strconv.FormatFloat(1/symbol.Last.InexactFloat64(), 'f', -1, 64),
			}
		}
	}
	return nil
}

func removeGateioDashFromSymbol(symbol string) string {
	return strings.ReplaceAll(strings.ToUpper(symbol), "_", "")
}
