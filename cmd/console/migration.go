package console

import (
	"fmt"

	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/dv-net/dv-merchant/pkg/migrations"

	"github.com/urfave/cli/v2"
)

func prepareMigrationCommands(currentAppVersion string) []*cli.Command {
	return []*cli.Command{
		{
			Name:        "up",
			Description: "up database schema",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name: "steps", Aliases: []string{"s"},
					Usage: "number of steps to migrate (all by default)",
				},
				&cli.BoolFlag{
					Name:  "disable-confirmations",
					Usage: "disable confirmation for migration",
				},
			},
			Action: func(ctx *cli.Context) error {
				conf, err := loadConfig(ctx.Args().Slice(), ctx.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				l := logger.New(currentAppVersion, conf.Log)
				mg, err := initMigrations(l, migrations.Config{
					DBDriver:            migrations.DBDriver(conf.Postgres.Engine()),
					DSN:                 conf.Postgres.DSN(),
					DisableConfirmation: ctx.Bool("disable-confirmations"),
				})
				if err != nil {
					return fmt.Errorf("failed to init migrations: %w", err)
				}
				return mg.Up(ctx.Context, ctx.Int("steps"))
			},
		}, // migrate.up
		{
			Name:        "down",
			Description: "rollback database schema",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name: "steps", Aliases: []string{"s"},
					Value: 1,
					Usage: "number of steps to migrate (1 by default)",
				},
				&cli.BoolFlag{
					Name:  "disable-confirmations",
					Usage: "disable confirmation for migration",
				},
			},
			Action: func(ctx *cli.Context) error {
				conf, err := loadConfig(ctx.Args().Slice(), ctx.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				l := logger.New(currentAppVersion, conf.Log)
				mg, err := initMigrations(l, migrations.Config{
					DBDriver:            migrations.DBDriver(conf.Postgres.Engine()),
					DSN:                 conf.Postgres.DSN(),
					DisableConfirmation: ctx.Bool("disable-confirmations"),
				})
				if err != nil {
					return fmt.Errorf("failed to init migrations: %w", err)
				}
				return mg.Down(ctx.Context, ctx.Int("steps"))
			},
		}, // migrate.down
		{
			Name:        "drop",
			Description: "drop database schema",
			Action: func(ctx *cli.Context) error {
				conf, err := loadConfig(ctx.Args().Slice(), ctx.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				l := logger.New(currentAppVersion, conf.Log)

				mg, err := initMigrations(l, migrations.Config{
					DBDriver:            migrations.DBDriver(conf.Postgres.Engine()),
					DSN:                 conf.Postgres.DSN(),
					DisableConfirmation: false,
				})
				if err != nil {
					return fmt.Errorf("failed to init migrations: %w", err)
				}
				return mg.Drop(ctx.Context)
			},
		}, // migrate.drop
		{
			Name:        "version",
			Description: "print current database schema version",
			Action: func(ctx *cli.Context) error {
				conf, err := loadConfig(ctx.Args().Slice(), ctx.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				l := logger.New(currentAppVersion, conf.Log)
				mg, err := initMigrations(l, migrations.Config{
					DBDriver:            migrations.DBDriver(conf.Postgres.Engine()),
					DSN:                 conf.Postgres.DSN(),
					DisableConfirmation: true,
				})
				if err != nil {
					return fmt.Errorf("failed to init migrations: %w", err)
				}

				ver, isDirty, err := mg.Version(ctx.Context)
				if err != nil {
					return fmt.Errorf("failed to get version: %w", err)
				}
				l.Info(fmt.Sprintf("version: %d, dirty: %t", ver, isDirty))
				return nil
			},
		}, // migrate.version
	}
}

type migrationsLogger struct {
	log logger.Logger
}

func (ml migrationsLogger) Infof(format string, args ...interface{}) {
	ml.log.Info(fmt.Sprintf(format, args...))
}

func initMigrations(l logger.Logger, conf migrations.Config) (*migrations.Migration, error) {
	return migrations.New(migrationsLogger{log: l}, conf)
}
