package exrate

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/shopspring/decimal"
)

type GateioSymbol struct {
	CurrencyPair string          `json:"currency_pair"`
	Last         decimal.Decimal `json:"last"`
}

type GateioResponse struct {
	Data []GateioSymbol `json:"data"`
}

func NewGateioFetcher(url string, httpClient *http.Client, log logger.Logger) IFetcher {
	return &gateioFetcher{url: url, httpClient: httpClient, log: log}
}

type gateioFetcher struct {
	url        string
	httpClient *http.Client
	log        logger.Logger
}

var _ IFetcher = (*gateioFetcher)(nil)

func (o *gateioFetcher) Source() string {
	return "gate"
}

func (o *gateioFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) (err error) {
	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, o.url, nil); err != nil {
		return err
	}
	var resp *http.Response
	if resp, err = o.httpClient.Do(req); err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := parseGateioResponse(resp.Body)
	if err != nil {
		return err
	}

	return filterGateioResponse(body, currencyFilter, out)
}

func parseGateioResponse(rc io.ReadCloser) (*GateioResponse, error) {
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	r := &GateioResponse{}
	if err := json.Unmarshal(b, &r.Data); err != nil {
		return nil, err
	}
	return r, nil
}

func filterGateioResponse(r *GateioResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error { //nolint:unparam
	for _, symbol := range r.Data {
		if s, ok := currencyFilter.symbols[removeGateioDashFromSymbol(symbol.CurrencyPair)]; ok {
			out <- ExRate{
				Source: "gate",
				From:   s.From,
				To:     s.To,
				Value:  symbol.Last.String(),
			}
			out <- ExRate{
				Source: "gate",
				From:   s.To,
				To:     s.From,
				Value:  strconv.FormatFloat(1/symbol.Last.InexactFloat64(), 'f', -1, 64),
			}
		}
	}
	return nil
}

func removeGateioDashFromSymbol(symbol string) string {
	return strings.ReplaceAll(strings.ToUpper(symbol), "_", "")
}
