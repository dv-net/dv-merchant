package rate

import "time"

type Option func(service *Service)

func WithDuration(duration time.Duration) Option {
	return func(s *Service) {
		s.rateTTL = duration
	}
}

func WithMaxLimit(maxValue int64) Option {
	return func(s *Service) {
		s.maxCounterValue = maxValue
	}
}
