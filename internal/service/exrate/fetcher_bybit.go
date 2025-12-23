//nolint:tagliatelle
package exrate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func NewBybitFetcher(url string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &bybitFetcher{url: url, httpClient: httpClient, log: log}
}

type bybitFetcher struct {
	url        string
	httpClient *http.Client
	log        logger.Logger
}

var _ IFetcher = (*bybitFetcher)(nil)

func (o *bybitFetcher) Source() string {
	return "bybit"
}

func (o *bybitFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) (err error) {
	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, o.url, http.NoBody); err != nil {
		o.log.Errorw("[EXRATE-BYBIT] failed to create request",
			"error", err,
			"url", o.url)
		return err
	}

	var resp *http.Response
	if resp, err = o.httpClient.Do(req); err != nil {
		o.log.Errorw("[EXRATE-BYBIT] http client error",
			"error", err,
			"url", o.url)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		o.log.Errorw("[EXRATE-BYBIT] failed to read response body",
			"error", err,
			"status_code", resp.StatusCode)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		o.log.Errorw("[EXRATE-BYBIT] non-OK HTTP status",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"raw_response", string(bodyBytes))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := parseBybitResponseBytes(bodyBytes)
	if err != nil {
		o.log.Errorw("[EXRATE-BYBIT] response parsing error",
			"error", err,
			"raw_response", string(bodyBytes),
			"status_code", resp.StatusCode)
		return err
	}

	if err := filterBybitResponse(body, currencyFilter, out); err != nil {
		o.log.Errorw("[EXRATE-BYBIT] failed to filter response",
			"error", err,
			"symbol_count", len(body.Result.List))
		return err
	}

	o.log.Infow("[EXRATE-BYBIT] successfully fetched exchange rates",
		"symbol_count", len(body.Result.List))

	return nil
}

func parseBybitResponseBytes(b []byte) (*BybitResponse, error) {
	r := &BybitResponse{}
	if err := json.Unmarshal(b, r); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
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
