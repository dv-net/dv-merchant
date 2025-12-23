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

type BinanceSymbol struct {
	Symbol string          `json:"symbol"`
	Price  decimal.Decimal `json:"price"`
}

type BinanceResponse []BinanceSymbol

func NewBinanceFetcher(url string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &binanceFetcher{url: url, httpClient: httpClient, log: log}
}

type binanceFetcher struct {
	url        string
	httpClient *http.Client
	log        logger.Logger
}

var _ IFetcher = (*binanceFetcher)(nil)

func (f *binanceFetcher) Source() string {
	return "binance"
}

func (f *binanceFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) (err error) {
	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, f.url, http.NoBody); err != nil {
		f.log.Errorw("[EXRATE-BINANCE] failed to create request",
			"error", err,
			"url", f.url)
		return err
	}

	var resp *http.Response
	if resp, err = f.httpClient.Do(req); err != nil {
		f.log.Errorw("[EXRATE-BINANCE] http client error",
			"error", err,
			"url", f.url)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// Read body once for both parsing and potential error logging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		f.log.Errorw("[EXRATE-BINANCE] failed to read response body",
			"error", err,
			"status_code", resp.StatusCode)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		f.log.Errorw("[EXRATE-BINANCE] non-OK HTTP status",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"raw_response", string(bodyBytes))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := parseBinanceResponseBytes(bodyBytes)
	if err != nil {
		f.log.Errorw("[EXRATE-BINANCE] response parsing error",
			"error", err,
			"raw_response", string(bodyBytes),
			"status_code", resp.StatusCode)
		return err
	}

	if err := filterBinanceResponse(body, currencyFilter, out); err != nil {
		f.log.Errorw("[EXRATE-BINANCE] failed to filter response",
			"error", err,
			"symbol_count", len(*body))
		return err
	}

	f.log.Infow("[EXRATE-BINANCE] successfully fetched exchange rates",
		"symbol_count", len(*body))

	return nil
}

func parseBinanceResponseBytes(b []byte) (*BinanceResponse, error) {
	r := &BinanceResponse{}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}
	return r, nil
}

func filterBinanceResponse(r *BinanceResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error { //nolint:unparam
	for _, symbol := range *r {
		if s, ok := currencyFilter.symbols[symbol.Symbol]; ok {
			out <- ExRate{
				Source: "binance",
				From:   s.From,
				To:     s.To,
				Value:  symbol.Price.String(),
			}
			out <- ExRate{
				Source: "binance",
				From:   s.To,
				To:     s.From,
				Value:  strconv.FormatFloat(1/symbol.Price.InexactFloat64(), 'f', -1, 64),
			}
		}
	}
	return nil
}
