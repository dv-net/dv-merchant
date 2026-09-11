package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5"
)

// migrationLockTimeout bounds how long the migration session waits for any lock,
// including golang-migrate's advisory lock (`SELECT pg_advisory_lock`). Without it
// a stale advisory lock left by a killed migration process, or a migration statement
// blocked behind a long-running transaction, hangs `migrate` indefinitely. With it
// the command fails fast with a clear Postgres error instead.
const migrationLockTimeout = "60s"

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

	// Applies to this session only; golang-migrate keeps using this same conn for
	// the advisory lock and every migration statement.
	if _, err := conn.ExecContext(ctx, "SET lock_timeout = '"+migrationLockTimeout+"'"); err != nil {
		return nil, fmt.Errorf("set lock_timeout: %w", err)
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
