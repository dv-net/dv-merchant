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
		return err
	}
	var resp *http.Response
	if resp, err = f.httpClient.Do(req); err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := parseBinanceResponse(resp.Body)
	if err != nil {
		return err
	}

	return filterBinanceResponse(body, currencyFilter, out)
}

func parseBinanceResponse(rc io.ReadCloser) (*BinanceResponse, error) {
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	r := &BinanceResponse{}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
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
