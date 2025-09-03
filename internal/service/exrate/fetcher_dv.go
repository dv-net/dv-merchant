package exrate

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/pkg/admin_gateway"
	"github.com/dv-net/dv-merchant/pkg/admin_gateway/responses"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

func NewDVFetcher(dvRate admin_gateway.IRates, log logger.Logger) IFetcher {
	return &dvFetcher{client: dvRate, log: log}
}

type dvFetcher struct {
	client admin_gateway.IRates
	log    logger.Logger
}

func (o *dvFetcher) Source() string {
	return "dv"
}

func (o *dvFetcher) Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	res, err := o.client.FetchAllRates(ctx)
	if err != nil {
		return fmt.Errorf("fetch all rates: %w", err)
	}

	return filterDvRates(res, currencyFilter, out)
}

func filterDvRates(r []responses.RatesResponse, currencyFilter CurrencyFilter, out chan<- ExRate) error {
	for _, rate := range r {
		if !currencyFilter.HasPair(CurrencyPair{From: rate.From, To: rate.To}) {
			continue
		}

		out <- ExRate{
			Source: rate.Source,
			From:   rate.From,
			To:     rate.To,
			Value:  rate.Value.String(),
		}

		out <- ExRate{
			Source: rate.Source,
			From:   rate.To,
			To:     rate.From,
			Value:  rate.Value.String(),
		}
	}

	return nil
}
