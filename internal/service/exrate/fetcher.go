package exrate

import (
	"context"
	"sync"
)

type IFetcher interface {
	Source() string // Source identifier
	Fetch(ctx context.Context, currencyFilter CurrencyFilter, out chan<- ExRate) error
}

type stubFetcher struct {
	source string
}

func (f *stubFetcher) Source() string {
	return f.source
}
func (f *stubFetcher) Fetch(_ context.Context, _ CurrencyFilter, _ chan<- ExRate) error {
	return nil
}

func (srv *service) fetchWorker(
	ctx context.Context, f IFetcher, in <-chan CurrencyFilter, out chan<- ExRate,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case cf := <-in:
			if err := f.Fetch(ctx, cf, out); err != nil {
				srv.logger.Error(
					"fetch rates", err,
					"source", f.Source(),
				)
			}
		}
	}
}
