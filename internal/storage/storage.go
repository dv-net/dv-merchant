package storage

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/pkg/database"
	"github.com/dv-net/dv-merchant/pkg/key_value"

	"github.com/redis/go-redis/v9"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IStorage interface {
	repos.IRepository
	PSQLConn() *pgxpool.Pool
	Close() error
}

type storage struct {
	repos.IRepository
	psql     *pgxpool.Pool
	keyValue key_value.IKeyValue
}

func InitStore(ctx context.Context, conf *config.Config) (IStorage, error) {
	var kv key_value.IKeyValue
	switch conf.KeyValue.Engine {
	case config.KeyValueEngineRedis:
		opts, err := redis.ParseURL(conf.Redis.URL())
		if err != nil {
			return nil, fmt.Errorf("parse redis url: %w", err)
		}
		redisClient := redis.NewClient(opts)
		if err = redisClient.Ping(ctx).Err(); err != nil {
			return nil, fmt.Errorf("ping redis: %w", err)
		}

		kv = key_value.NewRedisStorage(redisClient)
	case config.KeyValueEngineInMemory:
		kv = key_value.NewInMemory()
	default:
		return nil, fmt.Errorf("key_value engine '%s' is not supported", conf.KeyValue.Engine)
	}

	dbClient, err := database.NewPostgresClient(ctx, conf.Postgres)
	if err != nil {
		return nil, fmt.Errorf("failed to create database client: %w", err)
	}
	if err = dbClient.EnsureSchemaMigrationsReady(ctx); err != nil {
		return nil, fmt.Errorf("failed to create database schema migrations: %w", err)
	}

	return &storage{
		IRepository: repos.InitRepository(dbClient, kv),
		psql:        dbClient.DB,
		keyValue:    kv,
	}, nil
}

func (s storage) PSQLConn() *pgxpool.Pool { return s.psql }

func (s storage) Close() error {
	defer s.psql.Close()
	if err := s.keyValue.Close(); err != nil {
		return fmt.Errorf("key-value connection close: %w", err)
	}

	return nil
}
