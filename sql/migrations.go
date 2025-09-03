package sql

import (
	"embed"
)

//go:embed postgres/migrations/*.sql
var MigrationsPostgres embed.FS

type MigrationParameters struct {
	EmbedFs embed.FS
	Path    string
}

func PostgresMigrationParams() MigrationParameters {
	return MigrationParameters{
		Path:    "postgres/migrations",
		EmbedFs: MigrationsPostgres,
	}
}
