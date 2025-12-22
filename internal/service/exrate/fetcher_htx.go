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
	Status    string      `json:"status,omitempty"`
	Data      []HtxSymbol `json:"data,omitempty"`
	Timestamp int64       `json:"ts,omitempty"` //nolint:tagliatelle
}

func parseHtxResponse(rc io.ReadCloser) (*HtxResponse, error) {
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	r := &HtxResponse{}
	if err := json.Unmarshal(b, r); err != nil {
		return nil, err
	}
	return r, nil
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

func (o *htxFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	err := o.fetchWithClient(ctx, o.httpClient, currencyFilter, out)
	if err == nil {
		return nil // Success with direct connection
	}

	o.log.Debugw("direct request failed, trying proxies", "error", err)

	if len(o.proxies) == 0 {
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
			o.log.Debugw("failed to create proxy client", "proxy", proxyURL, "error", err)
			lastErr = err
			continue
		}

		err = o.fetchWithClient(ctx, client, currencyFilter, out)
		if err == nil {
			o.log.Debugw("request succeeded with proxy", "proxy", proxyURL)
			return nil // Success
		}

		o.log.Debugw("request failed with proxy, trying next", "proxy", proxyURL, "error", err)
		lastErr = err
	}

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

func (o *htxFetcher) fetchWithClient(ctx context.Context, client *http.Client, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := parseHtxResponse(resp.Body)
	if err != nil {
		return err
	}

	if body.Status != "ok" {
		o.log.Debugw(
			"currency exchange service response not OK",
			"status", body.Status,
		)
		return ErrInvalidResponseFromHtx
	}

	return filterHtxResponse(body, currencyFilter, out)
}

func filterHtxResponse(r *HtxResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error { //nolint:unparam
	for _, symbol := range r.Data {
		if s, ok := currencyFilter.symbols[strings.ToUpper(symbol.Symbol)]; ok {
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
