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

func (o *bybitFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error {
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

func (o *bybitFetcher) fetchWithClient(ctx context.Context, client *http.Client, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.url, http.NoBody)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := parseBybitResponse(resp.Body)
	if err != nil {
		return err
	}

	return filterBybitResponse(body, currencyFilter, out)
}

func parseBybitResponse(rc io.ReadCloser) (*BybitResponse, error) {
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	r := &BybitResponse{}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	return r, nil
}

func filterBybitResponse(r *BybitResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error { //nolint:unparam
	for _, symbol := range r.Result.List {
		if s, ok := currencyFilter.symbols[symbol.Symbol]; ok {
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
