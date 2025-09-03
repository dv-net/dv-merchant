package rate

import (
	"time"

	"github.com/dv-net/dv-merchant/pkg/key_value"

	"golang.org/x/net/context"
)

type Limiter interface {
	IsAllowedByKey(ctx context.Context, rateLimitKey string) bool
}

type Service struct {
	driver key_value.IKeyValue

	rateTTL         time.Duration
	maxCounterValue int64
}

func NewLimiter(driver key_value.IKeyValue, opts ...Option) *Service {
	s := &Service{driver: driver}
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Service) IsAllowedByKey(ctx context.Context, rateLimitKey string) bool {
	if err := s.driver.IncrementCounterWithLimit(ctx, rateLimitKey, s.maxCounterValue, s.rateTTL); err != nil {
		return false
	}

	return true
}
