//nolint:tagliatelle
package exrate

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
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

func NewBitgetFetcher(url string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &bitgetFetcher{url: url, httpClient: httpClient, log: log}
}

type bitgetFetcher struct {
	url        string
	httpClient *http.Client
	log        logger.Logger
}

func (o *bitgetFetcher) Source() string {
	return "bitget"
}

func (o *bitgetFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	var req *http.Request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.url, nil)
	if err != nil {
		return err
	}
	resp, err := o.httpClient.Do(req)
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
