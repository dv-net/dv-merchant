package kraken

import (
	"context"
	"strconv"
	"sync"
	"time"
)

type nonceSequence struct {
	sem  chan struct{}
	mu   sync.Mutex
	last uint64
}

var nonceRegistry sync.Map // apiKey -> *nonceSequence

func nonceFor(apiKey string) *nonceSequence {
	if seq, ok := nonceRegistry.Load(apiKey); ok {
		return seq.(*nonceSequence) //nolint:forcetypeassert
	}
	seq, _ := nonceRegistry.LoadOrStore(apiKey, &nonceSequence{sem: make(chan struct{}, 1)})
	return seq.(*nonceSequence) //nolint:forcetypeassert
}

func (s *nonceSequence) lock(ctx context.Context) error {
	select {
	case s.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *nonceSequence) unlock() {
	<-s.sem
}

func (s *nonceSequence) next() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	n := uint64(time.Now().UnixMicro()) //nolint:gosec
	if n <= s.last {
		n = s.last + 1
	}
	s.last = n
	return strconv.FormatUint(n, 10)
}
