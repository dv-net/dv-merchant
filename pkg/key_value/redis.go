package key_value

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisStorage struct {
	client *redis.Client
}

func NewRedisStorage(client *redis.Client) IKeyValue {
	return &redisStorage{
		client: client,
	}
}

func (rs *redisStorage) Get(ctx context.Context, key string) (KeyValueResult, error) {
	if v := rs.client.Get(ctx, key); v.Err() != nil && errors.Is(v.Err(), redis.Nil) {
		return nil, ErrEntryNotFound
	}

	kType, err := rs.client.Type(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("get key type: %w", err)
	}
	switch kType {
	case "hash":
		res, err := rs.client.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("get hash key from redis: %w", err)
		}
		return json.Marshal(res)
	case "string":
		return rs.client.Get(ctx, key).Bytes()
	default:
		return nil, fmt.Errorf("unsupported key type: %s", kType)
	}
}

func (rs *redisStorage) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var err error
	switch value.(type) {
	case string:
		_, err = rs.client.Set(ctx, key, value, expiration).Result()
	case map[string]interface{}:
		_, err = rs.client.HSet(ctx, key, value).Result()
		if err != nil {
			return fmt.Errorf("set hash: %w", err)
		}
		_, err = rs.client.Expire(ctx, key, expiration).Result()
		if err != nil {
			return fmt.Errorf("set hash ttl: %w", err)
		}
	}
	if err != nil {
		return fmt.Errorf("set key to redis: %w", err)
	}

	return err
}

func (rs *redisStorage) Delete(ctx context.Context, key string) error {
	if err := rs.client.Del(ctx, key).Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrEntryNotFound
		}

		return fmt.Errorf("delete entry from redis: %w", err)
	}

	return nil
}

func (rs *redisStorage) Keys(ctx context.Context, pattern string) ([]string, error) {
	return rs.client.Keys(ctx, pattern).Result()
}

func (rs *redisStorage) Close() error {
	return rs.client.Close()
}

func (rs *redisStorage) IncrementCounterWithLimit(ctx context.Context, key string, counterMax int64, ttl time.Duration) error {
	return rs.client.Watch(ctx, func(tx *redis.Tx) error {
		counter, err := tx.Get(ctx, key).Int64()
		if errors.Is(err, redis.Nil) {
			tx.Set(ctx, key, 1, ttl)
			return nil
		}

		if counter >= counterMax {
			return ErrCounterLimitExceeded
		}

		return tx.Incr(ctx, key).Err()
	})
}
