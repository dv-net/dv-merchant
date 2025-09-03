package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5"
)

func (s *Migration) postgresDriver(ctx context.Context) (database.Driver, error) {
	connConfig, err := pgx.ParseConfig(s.config.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse postgres database dsn failed: %w", err)
	}

	db, err := sql.Open("postgres", s.config.DSN)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("database connection: %w", err)
	}

	instance, err := postgres.WithConnection(
		ctx, conn,
		&postgres.Config{
			MigrationsTable: "schema_migrations",
			DatabaseName:    connConfig.Database,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("init postgres instance error: %w", err)
	}

	return instance, nil
}
