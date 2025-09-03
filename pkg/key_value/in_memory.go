package key_value

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type inMemory struct {
	client   *ttlcache.Cache[string, []byte]
	counters *ttlcache.Cache[string, *atomic.Int64]
}

func NewInMemory() IKeyValue {
	return &inMemory{
		ttlcache.New[string, []byte](),
		ttlcache.New[string, *atomic.Int64](),
	}
}

func (im *inMemory) Get(_ context.Context, key string) (KeyValueResult, error) {
	entry := im.client.Get(key)
	if entry == nil {
		return nil, ErrEntryNotFound
	}

	return entry.Value(), nil
}

func (im *inMemory) Set(_ context.Context, key string, value interface{}, expiration time.Duration) error {
	if strVal, ok := value.([]byte); ok {
		_ = im.client.Set(key, strVal, expiration)
		return nil
	}
	if strVal, ok := value.(string); ok {
		_ = im.client.Set(key, []byte(strVal), expiration)
		return nil
	}

	return fmt.Errorf("unsupported value type: %T", value)
}

func (im *inMemory) Delete(_ context.Context, key string) error {
	im.client.Delete(key)
	return nil
}

func (im *inMemory) IncrementCounterWithLimit(_ context.Context, key string, counterMax int64, ttl time.Duration) error {
	res := im.counters.Get(key)
	if res == nil {
		counter := &atomic.Int64{}
		counter.Add(1)
		im.counters.Set(key, counter, ttl)
		return nil
	}

	val := res.Value()
	if val.Load() >= counterMax {
		return ErrCounterLimitExceeded
	}

	val.Add(1)
	return nil
}

func (im *inMemory) Close() error {
	return nil
}
