package exrate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

func NewKucoinFetcher(url string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &kucoinFetcher{url: url, httpClient: httpClient, log: log}
}

type kucoinFetcher struct {
	url        string
	httpClient *http.Client
	log        logger.Logger
}

var _ IFetcher = (*kucoinFetcher)(nil)

func (o *kucoinFetcher) Source() string {
	return "kucoin"
}

func (o *kucoinFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) (err error) {
	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, o.url, nil); err != nil {
		o.log.Errorw("[EXRATE-KUCOIN] failed to create request", "error", err, "url", o.url)
		return err
	}

	var resp *http.Response
	if resp, err = o.httpClient.Do(req); err != nil {
		o.log.Errorw("[EXRATE-KUCOIN] http client error", "error", err, "url", o.url)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		o.log.Errorw("[EXRATE-KUCOIN] failed to read response body", "error", err, "status_code", resp.StatusCode)
		return err
	}

	body, err := parseKucoinResponseBytes(bodyBytes)
	if err != nil {
		o.log.Errorw("[EXRATE-KUCOIN] response parsing error",
			"error", err,
			"raw_response", string(bodyBytes),
			"status_code", resp.StatusCode)
		return err
	}

	if body.Code != "200000" {
		o.log.Errorw(
			"[EXRATE-KUCOIN] currency exchange service response not OK",
			"raw_response", string(bodyBytes),
			"status", body.Code,
		)
		return ErrInvalidResponseFromKucoin
	}

	if err := filterKucoinResponse(body, currencyFilter, out); err != nil {
		o.log.Errorw("[EXRATE-KUCOIN] failed to filter response",
			"error", err,
			"ticker_count", len(body.Data.Ticker))
		return err
	}

	return nil
}

func parseKucoinResponseBytes(b []byte) (*KucoinResponse, error) {
	r := &KucoinResponse{}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
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
