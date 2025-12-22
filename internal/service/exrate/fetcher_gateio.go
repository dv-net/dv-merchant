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

func (o *gateioFetcher) fetchWithClient(ctx context.Context, client *http.Client, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := parseGateioResponse(resp.Body)
	if err != nil {
		return err
	}

	return filterGateioResponse(body, currencyFilter, out)
}

func parseGateioResponse(rc io.ReadCloser) (*GateioResponse, error) {
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	r := &GateioResponse{}
	if err := json.Unmarshal(b, &r.Data); err != nil {
		return nil, err
	}
	return r, nil
}

func filterGateioResponse(r *GateioResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error { //nolint:unparam
	for _, symbol := range r.Data {
		if s, ok := currencyFilter.symbols[removeGateioDashFromSymbol(symbol.CurrencyPair)]; ok {
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
