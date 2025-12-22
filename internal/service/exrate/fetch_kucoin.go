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

func (o *kucoinFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error {
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

func (o *kucoinFetcher) fetchWithClient(ctx context.Context, client *http.Client, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := parseKucoinResponse(resp.Body)
	if err != nil {
		return err
	}

	if body.Code != "200000" {
		o.log.Debugw(
			"currency exchange service response not OK",
			"status", body.Code,
		)
		return ErrInvalidResponseFromKucoin
	}

	return filterKucoinResponse(body, currencyFilter, out)
}

func parseKucoinResponse(rc io.ReadCloser) (*KucoinResponse, error) {
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	r := &KucoinResponse{}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	return r, nil
}

func filterKucoinResponse(r *KucoinResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	for _, symbol := range r.Data.Ticker {
		if s, ok := currencyFilter.symbols[removeDashFromSymbol(symbol.Symbol)]; ok {
			floatValue, err := strconv.ParseFloat(symbol.Last, 64)
			if err != nil {
				return err
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
