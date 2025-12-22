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

func parseOkxResponse(rc io.ReadCloser) (*OkxResponse, error) {
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	r := &OkxResponse{}
	if err := json.Unmarshal(b, r); err != nil {
		return nil, err
	}
	return r, nil
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
	err := o.fetchAllCurrencies(ctx, o.httpClient, currencyFilter, out)
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

		err = o.fetchAllCurrencies(ctx, client, currencyFilter, out)
		if err == nil {
			o.log.Debugw("request succeeded with proxy", "proxy", proxyURL)
			return nil // Success
		}

		o.log.Debugw("request failed with proxy, trying next", "proxy", proxyURL, "error", err)
		lastErr = err
	}

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

func (o *okxFetcher) fetchAllCurrencies(ctx context.Context, client *http.Client, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	mergeCh := make(chan ExRate, 3)
	ccys := []string{"USDT", "USDC", "BTC"}
	var wg sync.WaitGroup
	wg.Add(len(ccys))

	errCh := make(chan error, len(ccys))

	for _, currency := range ccys {
		go func(currency string) {
			defer wg.Done()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.url, nil)
			if err != nil {
				errCh <- err
				return
			}
			q := req.URL.Query()
			q.Add("quoteCcy", currency)
			req.URL.RawQuery = q.Encode()

			resp, err := client.Do(req)
			if err != nil {
				errCh <- err
				return
			}
			defer func() { _ = resp.Body.Close() }()
			body, err := parseOkxResponse(resp.Body)
			if err != nil {
				errCh <- err
				return
			}

			if body.Code != "0" {
				o.log.Debugw("currency exchange service response not OK", "status", body.Message)
				errCh <- ErrInvalidResponseFromOkx
				return
			}
			_ = filterOkxResponse(body, currencyFilter, mergeCh)
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

	// Check if any errors occurred
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func filterOkxResponse(r *OkxResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	for _, symbol := range r.Data {
		if s, ok := currencyFilter.symbols[removeDashFromSymbol(symbol.InstID)]; ok {
			floatValue, err := strconv.ParseFloat(symbol.IdxPx, 64)
			if err != nil {
				return err
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
