package database

import "context"

type DBClient interface {
	Close() error
	EnsureSchemaMigrationsReady(ctx context.Context) error
}

type DB struct{}
