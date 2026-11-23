package key_value

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
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
	switch v := value.(type) {
	case string:
		im.client.Set(key, []byte(v), expiration)
	case []byte:
		im.client.Set(key, v, expiration)
	case map[string]interface{}:
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("marshal map value: %w", err)
		}
		im.client.Set(key, data, expiration)
	default:
		return fmt.Errorf("unsupported value type: %T", value)
	}
	return nil
}

func (im *inMemory) Delete(_ context.Context, key string) error {
	// Like Redis DEL: removes the key whatever it holds, including counters.
	im.client.Delete(key)
	im.counters.Delete(key)
	return nil
}

func (im *inMemory) IncrementCounterWithLimit(_ context.Context, key string, counterMax int64, ttl time.Duration) error {
	// GetOrSet + CAS loop: concurrent callers can neither create two counters nor overshoot the limit.
	item, _ := im.counters.GetOrSet(key, &atomic.Int64{}, ttlcache.WithTTL[string, *atomic.Int64](ttl))
	val := item.Value()
	for {
		cur := val.Load()
		if cur >= counterMax {
			return ErrCounterLimitExceeded
		}
		if val.CompareAndSwap(cur, cur+1) {
			return nil
		}
	}
}

func (im *inMemory) Keys(_ context.Context, pattern string) ([]string, error) {
	keys := make([]string, 0)

	for key := range im.client.Items() {
		matched, err := filepath.Match(pattern, key)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern: %w", err)
		}
		if matched {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

func (im *inMemory) Close() error {
	return nil
}
