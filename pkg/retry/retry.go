package retry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Retry represents a retry mechanism
type Retry struct {
	ctx         context.Context
	logger      logger
	maxAttempts int
	policy      Policy
	delay       time.Duration
	onFailedFn  func()
	onSuccessFn func()
}

// New creates a new Retry instance
//
// maxAttempts: the maximum number of attempts. Default is 5
//
// policy: the retry policy. Default is PolicyBackoff
func New(opts ...Option) *Retry {
	r := &Retry{
		maxAttempts: 5,
		policy:      PolicyBackoff,
		delay:       1 * time.Second,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Do - performs the retry mechanism
func (r *Retry) Do(fn func() error) error {
	if fn == nil {
		return fmt.Errorf("retry function cannot be nil")
	}

	var err error
	switch r.policy {
	case PolicyLinear:
		err = r.linearRetry(fn)
	case PolicyBackoff:
		err = r.backoffRetry(fn)
	case PolicyInfinite:
		err = r.infiniteRetry(fn)
	default:
		err = fmt.Errorf("unsupported retry policy")
	}

	if err == nil && r.onSuccessFn != nil {
		r.onSuccessFn()
	}

	return err
}

// linearRetry - performs a linear retry mechanism
func (r *Retry) linearRetry(fn func() error) error {
	for attempt := 1; attempt <= r.maxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		if r.onFailedFn != nil {
			r.onFailedFn()
		}

		if errors.Is(err, ErrExit) {
			return err
		}

		if attempt < r.maxAttempts {
			if r.logger != nil {
				r.logger.Infof("linear retry attempt %d failed, retrying in %s...", attempt, r.delay)
			}
			time.Sleep(r.delay)
		}
	}
	return fmt.Errorf("linear retry failed after %d attempts", r.maxAttempts)
}

// backoffRetry - performs a backoff retry mechanism
func (r *Retry) backoffRetry(fn func() error) error {
	for attempt := 1; attempt <= r.maxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		if r.onFailedFn != nil {
			r.onFailedFn()
		}

		if errors.Is(err, ErrExit) {
			return err
		}

		if attempt < r.maxAttempts {
			delay := r.delay * (1 << (attempt - 1)) // Increase a delay of 2 times by each attempt
			if r.logger != nil {
				r.logger.Infof("backoff retry attempt %d failed, retrying in %s...", attempt, delay)
			}
			time.Sleep(delay)
		}
	}
	return fmt.Errorf("backoff retry failed after %d attempts", r.maxAttempts)
}

func (r *Retry) infiniteRetry(fn func() error) error {
	if r.ctx == nil {
		return fmt.Errorf("infinite retry cannot be initialized without ctx")
	}

	resCh := make(chan error, 1)
	go func() {
		defer close(resCh)
		for {
			select {
			case <-r.ctx.Done():
				return
			default:
				err := fn()
				if err == nil {
					return
				}

				if r.onFailedFn != nil {
					r.onFailedFn()
				}

				if errors.Is(err, ErrExit) {
					resCh <- err
					return
				}

				if r.logger != nil {
					r.logger.Infof("infinite retry attempt failed, retrying in %s...", r.delay)
				}
				time.Sleep(r.delay)
			}
		}
	}()

	return <-resCh
}
