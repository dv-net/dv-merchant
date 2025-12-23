//nolint:tagliatelle
package exrate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/dv-net/dv-merchant/pkg/logger"
)

// Rate Limit: 20 requests per 2 seconds
// https://www.okx.com/docs-v5/en/#public-data-rest-api-get-index-tickers
// https://www.okx.com/api/v5/market/index-tickers?quoteCcy=USDT

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

func NewOkxFetcher(url string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &okxFetcher{url: url, httpClient: httpClient, log: log}
}

type okxFetcher struct {
	url        string
	httpClient *http.Client
	log        logger.Logger
}

func (o *okxFetcher) Source() string {
	return "okx"
}

func (o *okxFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	mergeCh := make(chan ExRate, 100)
	errCh := make(chan error, 3)

	ccys := []string{"USDT", "USDC", "BTC"}
	var wg sync.WaitGroup
	wg.Add(len(ccys))

	for _, currency := range ccys {
		go func(currency string) {
			defer wg.Done()
			if err := o.fetchForCurrency(ctx, currency, currencyFilter, mergeCh); err != nil {
				o.log.Errorw("[EXRATE-OKX] failed to fetch for currency",
					"error", err,
					"currency", currency)
				errCh <- fmt.Errorf("currency %s: %w", currency, err)
			} else {
				o.log.Infow("[EXRATE-OKX] successfully fetched for currency",
					"currency", currency)
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
			"errors", errs)
		return fmt.Errorf("all OKX fetches failed: %v", errs)
	}

	if len(errs) > 0 {
		o.log.Warnw("[EXRATE-OKX] partial fetch failures",
			"failed_count", len(errs),
			"success_count", len(ccys)-len(errs),
			"errors", errs)
	}

	return nil
}

func (o *okxFetcher) fetchForCurrency(ctx context.Context, currency string, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("quoteCcy", currency)
	req.URL.RawQuery = q.Encode()

	o.log.Debugw("[EXRATE-OKX] making request",
		"url", req.URL.String(),
		"currency", currency)

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http client error: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read body once for both parsing and error logging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	body, err := parseOkxResponseBytes(bodyBytes)
	if err != nil {
		return fmt.Errorf("response parsing error: %w, raw: %s", err, string(bodyBytes))
	}

	if body.Code != "0" {
		// Changed from Debugw to proper error return
		return fmt.Errorf("%w: code=%s, msg=%s", ErrInvalidResponseFromOkx, body.Code, body.Message)
	}

	if err := filterOkxResponse(body, currencyFilter, out); err != nil {
		return fmt.Errorf("failed to filter response: %w", err)
	}

	o.log.Debugw("[EXRATE-OKX] successfully processed response",
		"currency", currency,
		"symbol_count", len(body.Data))

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

	processedCount := 0
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
			processedCount++
		}
	}

	return nil
}

func removeDashFromSymbol(symbol string) string {
	return strings.ReplaceAll(symbol, "-", "")
}
