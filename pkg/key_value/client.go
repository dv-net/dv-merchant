package key_value

import (
	"context"
	"time"
)

type KeyValueResult []byte

func (o KeyValueResult) String() string {
	return string(o)
}

func (o KeyValueResult) Bytes() []byte {
	return o
}

type IKeyValue interface {
	Get(ctx context.Context, key string) (KeyValueResult, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	IncrementCounterWithLimit(ctx context.Context, key string, counterMax int64, ttl time.Duration) error
	Close() error
}

var (
	_ IKeyValue = (*redisStorage)(nil)
	_ IKeyValue = (*inMemory)(nil)
)
