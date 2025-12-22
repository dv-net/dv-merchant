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

	"github.com/dv-net/dv-merchant/pkg/logger"
)

var ErrInvalidResponseFromBitget = errors.New("invalid response from bitget")

type BitgetSymbol struct {
	Symbol       string  `json:"symbol"`
	High24h      float64 `json:"high24h,string"`
	Open         float64 `json:"open,string"`
	Low24h       float64 `json:"low24h,string"`
	LastPr       float64 `json:"lastPr,string"`
	QuoteVolume  float64 `json:"quoteVolume,string"`
	BaseVolume   float64 `json:"baseVolume,string"`
	UsdtVolume   float64 `json:"usdtVolume,string"`
	BidPr        float64 `json:"bidPr,string"`
	AskPr        float64 `json:"askPr,string"`
	BidSz        float64 `json:"bidSz,string"`
	AskSz        float64 `json:"askSz,string"`
	OpenUtc      float64 `json:"openUtc,string"`
	TS           int64   `json:"ts,string"`
	ChangeUtc24h float64 `json:"changeUtc24h,string"`
	Change24h    float64 `json:"change24h,string"`
}

type BitgetResponse struct {
	Code        string         `json:"code,omitempty"`
	Msg         string         `json:"msg,omitempty"`
	RequestTime int64          `json:"requestTime,omitempty"`
	Data        []BitgetSymbol `json:"data,omitempty"`
}

func parseBitgetResponse(rc io.ReadCloser) (*BitgetResponse, error) {
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	r := &BitgetResponse{}
	if err := json.Unmarshal(b, r); err != nil {
		return nil, err
	}
	return r, nil
}

func NewBitgetFetcher(url string, proxies []string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &bitgetFetcher{url: url, proxies: proxies, httpClient: httpClient, log: log}
}

type bitgetFetcher struct {
	url        string
	httpClient *http.Client
	proxies    []string
	log        logger.Logger
}

func (o *bitgetFetcher) Source() string {
	return "bitget"
}

func (o *bitgetFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error {
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

func (o *bitgetFetcher) createProxyClient(proxyURL string) (*http.Client, error) {
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

func (o *bitgetFetcher) fetchWithClient(ctx context.Context, client *http.Client, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := parseBitgetResponse(resp.Body)
	if err != nil {
		return err
	}

	if body.Code != "00000" {
		o.log.Debugw(
			"currency exchange service response not OK",
			"status", body.Msg,
		)
		return ErrInvalidResponseFromBitget
	}

	return filterBitgetResponse(body, currencyFilter, out)
}

func filterBitgetResponse(r *BitgetResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error { //nolint:unparam
	for _, symbol := range r.Data {
		if s, ok := currencyFilter.symbols[strings.ToUpper(symbol.Symbol)]; ok {
			askRounded := strconv.FormatFloat(roundToSixDecimalPlaces(symbol.AskPr), 'f', 6, 64)
			out <- ExRate{
				Source: "bitget",
				From:   s.From,
				To:     s.To,
				Value:  askRounded,
			}
			out <- ExRate{
				Source: "bitget",
				From:   s.To,
				To:     s.From,
				Value:  strconv.FormatFloat(roundToSixDecimalPlaces(1/symbol.AskPr), 'f', 6, 64),
			}
		}
	}
	return nil
}
