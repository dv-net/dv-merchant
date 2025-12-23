package exrate

import (
	"context"
	"encoding/json"
	"fmt"
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
		o.log.Errorw("[EXRATE-GATE] failed to create request",
			"error", err,
			"url", o.url)
		return err
	}

	var resp *http.Response
	if resp, err = o.httpClient.Do(req); err != nil {
		o.log.Errorw("[EXRATE-GATE] http client error",
			"error", err,
			"url", o.url)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		o.log.Errorw("[EXRATE-GATE] failed to read response body",
			"error", err,
			"status_code", resp.StatusCode)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		o.log.Errorw("[EXRATE-GATE] non-OK HTTP status",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"raw_response", string(bodyBytes))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := parseGateioResponseBytes(bodyBytes)
	if err != nil {
		o.log.Errorw("[EXRATE-GATE] response parsing error",
			"error", err,
			"raw_response", string(bodyBytes),
			"status_code", resp.StatusCode)
		return err
	}

	if err := filterGateioResponse(body, currencyFilter, out); err != nil {
		o.log.Errorw("[EXRATE-GATE] failed to filter response",
			"error", err,
			"symbol_count", len(body.Data))
		return err
	}

	o.log.Infow("[EXRATE-GATE] successfully fetched exchange rates",
		"symbol_count", len(body.Data))

	return nil
}

func parseGateioResponseBytes(b []byte) (*GateioResponse, error) {
	r := &GateioResponse{}
	if err := json.Unmarshal(b, &r.Data); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
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
