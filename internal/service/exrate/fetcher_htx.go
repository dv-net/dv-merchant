//nolint:tagliatelle
package exrate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/dv-net/dv-merchant/pkg/logger"
)

// https://api.huobi.pro/market/detail?symbol=btcusdt
// https://api.huobi.pro/market/tickers

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
	Status  string      `json:"status,omitempty"`
	Data    []HtxSymbol `json:"data,omitempty"`
	ErrCode string      `json:"err-code,omitempty"`
	ErrMsg  string      `json:"err-msg,omitempty"`
}

func NewHtxFetcher(url string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &htxFetcher{url: url, httpClient: httpClient, log: log}
}

type htxFetcher struct {
	url        string
	httpClient *http.Client
	log        logger.Logger
}

func (o *htxFetcher) Source() string {
	return "htx"
}

func (o *htxFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	var req *http.Request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.url, nil)
	if err != nil {
		o.log.Errorw("[EXRATE-HTX] failed to create request",
			"error", err,
			"url", o.url)
		return err
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		o.log.Errorw("[EXRATE-HTX] http client error",
			"error", err,
			"url", o.url)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		o.log.Errorw("[EXRATE-HTX] failed to read response body",
			"error", err,
			"status_code", resp.StatusCode)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		o.log.Errorw("[EXRATE-HTX] non-OK HTTP status",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"raw_response", string(bodyBytes))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := parseHtxResponseBytes(bodyBytes)
	if err != nil {
		o.log.Errorw("[EXRATE-HTX] response parsing error",
			"error", err,
			"raw_response", string(bodyBytes),
			"status_code", resp.StatusCode)
		return err
	}

	if body.Status != "ok" {
		o.log.Errorw(
			"[EXRATE-HTX] currency exchange service response not OK",
			"status", body.Status,
			"err_code", body.ErrCode,
			"err_msg", body.ErrMsg,
			"raw_response", string(bodyBytes))
		return fmt.Errorf("%w: %s - %s", ErrInvalidResponseFromHtx, body.ErrCode, body.ErrMsg)
	}

	if err := filterHtxResponse(body, currencyFilter, out); err != nil {
		o.log.Errorw("[EXRATE-HTX] failed to filter response",
			"error", err,
			"symbol_count", len(body.Data))
		return err
	}

	o.log.Infow("[EXRATE-HTX] successfully fetched exchange rates",
		"symbol_count", len(body.Data))

	return nil
}

func parseHtxResponseBytes(b []byte) (*HtxResponse, error) {
	r := &HtxResponse{}
	if err := json.Unmarshal(b, r); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}
	return r, nil
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
