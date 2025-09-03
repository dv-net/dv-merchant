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
	mergeCh := make(chan ExRate, 3)
	ccys := []string{"USDT", "USDC", "BTC"}
	var wg sync.WaitGroup
	wg.Add(len(ccys))
	for _, currency := range ccys {
		go func(currency string) {
			defer wg.Done()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.url, nil)
			if err != nil {
				return
			}
			q := req.URL.Query()
			q.Add("quoteCcy", currency)
			req.URL.RawQuery = q.Encode()

			resp, err := o.httpClient.Do(req)
			if err != nil {
				return
			}
			defer func() { _ = resp.Body.Close() }()
			body, err := parseOkxResponse(resp.Body)
			if err != nil {
				return
			}

			if body.Code != "0" {
				o.log.Debug("currency exchange service response not OK", "status", body.Message)
				return
			}
			_ = filterOkxResponse(body, currencyFilter, mergeCh)
		}(currency)
	}

	go func() {
		wg.Wait()
		close(mergeCh)
	}()

	for rate := range mergeCh {
		out <- rate
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
