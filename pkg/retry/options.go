package retry

import (
	"context"
	"time"
)

type Option func(*Retry)

// WithLogger sets the logger for the Retry instance
func WithLogger(logger logger) Option {
	return func(r *Retry) {
		r.logger = logger
	}
}

// WithMaxAttempts sets the maximum number of attempts
func WithMaxAttempts(maxAttempts int) Option {
	return func(r *Retry) {
		r.maxAttempts = maxAttempts
	}
}

// WithPolicy sets the retry policy
func WithPolicy(policy Policy) Option {
	return func(r *Retry) {
		r.policy = policy
	}
}

// WithDelay sets the delay between retries
func WithDelay(delay time.Duration) Option {
	return func(r *Retry) {
		r.delay = delay
	}
}

// WithContext sets the ctx for Infinite policy retry
func WithContext(ctx context.Context) Option {
	return func(r *Retry) {
		r.ctx = ctx
	}
}

// WithOnFailedFn sets the function to be called on failure
func WithOnFailedFn(fn func()) Option {
	return func(r *Retry) {
		r.onFailedFn = fn
	}
}

// WithOnSuccessFn sets the function to be called on success
func WithOnSuccessFn(fn func()) Option {
	return func(r *Retry) {
		r.onSuccessFn = fn
	}
}

// SetLogger sets the logger for the Retry instance
func (r *Retry) SetLogger(logger logger) *Retry {
	r.logger = logger
	return r
}

// SetMaxAttempts sets the maximum number of attempts
func (r *Retry) SetMaxAttempts(maxAttempts int) *Retry {
	r.maxAttempts = maxAttempts
	return r
}

// SetPolicy sets the retry policy
func (r *Retry) SetPolicy(policy Policy) *Retry {
	r.policy = policy
	return r
}

// SetDelay sets the delay between retries
func (r *Retry) SetDelay(delay time.Duration) *Retry {
	r.delay = delay
	return r
}

// SetContext sets the ctx for Infinite policy retry
func (r *Retry) SetContext(ctx context.Context) *Retry {
	r.ctx = ctx
	return r
}

// SetOnFailedFn sets the function to be called on failure
func (r *Retry) SetOnFailedFn(fn func()) *Retry {
	r.onFailedFn = fn
	return r
}

// SetOnSuccessFn sets the function to be called on success
func (r *Retry) SetOnSuccessFn(fn func()) *Retry {
	r.onSuccessFn = fn
	return r
}
