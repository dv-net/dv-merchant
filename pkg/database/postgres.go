package database

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresClient struct {
	DB *pgxpool.Pool
}

func NewPostgresClient(ctx context.Context, dbConf config.PostgresDB) (*PostgresClient, error) {
	cfg, err := pgxpool.ParseConfig(dbConf.DSN())
	if err != nil {
		return nil, err
	}

	cfg.MinConns = dbConf.MinOpenConns
	cfg.MaxConns = dbConf.MaxOpenConns

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &PostgresClient{DB: pool}, nil
}

func (c *PostgresClient) EnsureSchemaMigrationsReady(ctx context.Context) error {
	res, err := c.DB.Query(ctx, "SELECT sm.version, sm.dirty FROM schema_migrations sm WHERE sm.dirty=true LIMIT 1;")
	if err != nil {
		return fmt.Errorf("checking schema migrations: %w", err)
	}
	defer res.Close()

	if res.Next() {
		return fmt.Errorf("database schema is dirty")
	}

	return nil
}

func (c *PostgresClient) Close() error {
	if c.DB != nil {
		c.DB.Close()
	}
	return nil
}
