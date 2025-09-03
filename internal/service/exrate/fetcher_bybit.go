//nolint:tagliatelle
package exrate

import (
	"context"
	"encoding/json"
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
		return err
	}
	var resp *http.Response
	if resp, err = o.httpClient.Do(req); err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := parseBybitResponse(resp.Body)
	if err != nil {
		return err
	}

	return filterBybitResponse(body, currencyFilter, out)
}

func parseBybitResponse(rc io.ReadCloser) (*BybitResponse, error) {
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	r := &BybitResponse{}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
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
