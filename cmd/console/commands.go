package console

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/currency"
	"github.com/dv-net/dv-merchant/internal/service/eproxy"
	"github.com/dv-net/dv-merchant/internal/service/exrate"
	"github.com/dv-net/dv-merchant/internal/service/transactions"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/dv-net/xconfig"
	"github.com/goccy/go-yaml"

	"github.com/dv-net/dv-merchant/internal/app"
	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/seed"
	"github.com/dv-net/dv-merchant/internal/service/permission"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/pkg/migrations"
	sqlpostgres "github.com/dv-net/dv-merchant/sql/postgres"

	"connectrpc.com/connect"

	"github.com/urfave/cli/v2"

	epr "github.com/dv-net/dv-proto/go/eproxy"
)

const (
	defaultConfigPath = "configs/config.yaml"
	envPrefix         = "MERCHANT"
)

func InitCommands(currentAppVersion, commitHash string) []*cli.Command {
	return []*cli.Command{
		{
			Name:        "start",
			Description: "DV backend server",
			Flags:       []cli.Flag{cfgPathsFlag()},
			Action: func(ctx *cli.Context) error {
				conf, err := loadConfig(ctx.Args().Slice(), ctx.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				PrintBanner(currentAppVersion)
				l := logger.New(currentAppVersion, conf.Log)
				l.Info("Logger Init")
				return app.Run(ctx.Context, conf, l, currentAppVersion, commitHash)
			},
		}, // Start
		{
			Name:        "version",
			Description: "print DV backend server version",
			Action: func(_ *cli.Context) error {
				_, _ = fmt.Fprintln(os.Stdout, currentAppVersion)
				return nil
			},
		}, // version
		{
			Name:        "seed",
			Description: "DV database seed",
			Flags: []cli.Flag{
				cfgPathsFlag(),
				&cli.StringFlag{
					Name: "mode", Aliases: []string{"m"},
					Usage: "Seed mode: {up|down}",
					Value: "up",
					Action: func(_ *cli.Context, s string) error {
						switch strings.ToLower(s) {
						case string(seed.ModeUp), string(seed.ModeDown):
							return nil
						}
						return fmt.Errorf("invalid seed db mode: %s", s)
					},
				},
				&cli.StringFlag{
					Name: "name", Aliases: []string{"n"},
					Usage: "Name of seed",
				},
				&cli.PathFlag{
					Name: "file", Aliases: []string{"f"},
					Usage: "Full path to seed file",
				},
			},
			UsageText: `
If the "path" parameter is set, then the "driver" and "mode" parameters are ignored.
By default, the "mode" parameter is set to "up", and the "driver" parameter is set to "postgres".
If the parameter is not specified, then all scripts of the specified driver are executed in the specified mode.`,
			Action: func(ctx *cli.Context) error {
				conf, err := loadConfig(ctx.Args().Slice(), ctx.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				fmt.Printf("%+v\n", conf.Seed.Base)
				seeder, err := seed.NewSeeder(ctx.Context, conf, sqlpostgres.SeedsFs)
				if err != nil {
					return fmt.Errorf("failed to init: %w", err)
				}
				return seeder.Run(ctx.Context, &seed.Options{
					Mode: seed.Mode(ctx.String("mode")),
					Name: ctx.String("name"),
					File: ctx.String("file"),
				})
			},
		}, // seed
		{
			Name:        "config",
			Description: "validate, gen envs and flags for config",
			Subcommands: prepareConfigCommands(),
		}, // config
		{
			Name:        "migrate",
			Description: "migration database schema",
			Flags:       []cli.Flag{cfgPathsFlag()},
			Subcommands: prepareMigrationCommands(currentAppVersion),
		}, // migrate
		{
			Name:        "permission",
			Description: "Access rights management.",
			Flags:       []cli.Flag{cfgPathsFlag()},
			Subcommands: preparePermissionCommands(currentAppVersion),
		}, // permission
		{
			Name:        "transactions",
			Description: "Transactions management",
			Flags:       []cli.Flag{cfgPathsFlag()},
			Subcommands: prepareTransactionsCommands(currentAppVersion),
		}, // transactions
		{
			Name:        "users",
			Description: "Users management",
			Flags:       []cli.Flag{cfgPathsFlag()},
			Subcommands: prepareUsersCommands(),
		}, // user
	}
}

func preparePermissionCommands(currentAppVersion string) []*cli.Command {
	return []*cli.Command{
		{
			Name:        "load",
			Description: "load permission policies from CSV file",
			Flags: []cli.Flag{
				&cli.PathFlag{
					Name:  "file",
					Usage: "path to casbin rbac policies CSV file",
					Value: "configs/rbac_policies.csv",
				},
			},
			Action: func(ctx *cli.Context) error {
				conf, err := loadConfig(ctx.Args().Slice(), ctx.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				lg := logger.New(currentAppVersion, conf.Log)
				if err != nil {
					return fmt.Errorf("init logger failed: %w", err)
				}

				st, err := storage.InitStore(ctx.Context, conf)
				if err != nil {
					return fmt.Errorf("storage init: %w", err)
				}
				defer func() {
					if storageCloseErr := st.Close(); storageCloseErr != nil {
						lg.Errorw("storage close error", "error", storageCloseErr)
					}
				}()
				srv, err := permission.New(conf, st.PSQLConn())
				if err != nil {
					return fmt.Errorf("init permission service failed: %w", err)
				}
				if err := srv.LoadPolicies(ctx.String("file")); err != nil {
					return fmt.Errorf("load permission policies failed: %w", err)
				}
				return nil
			},
		}, // permission.load
		{
			Name:        "user",
			Description: "user access rights management",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "id",
					Usage: "user id",
				},
				&cli.StringFlag{
					Name:    "role",
					Aliases: []string{"r"},
					Usage:   "permission role",
				},
			},
			Subcommands: prepareUserPermissionCommands(currentAppVersion),
		}, // permission.user
		{
			Name:        "clear",
			Description: "delete all permission policies",
			Action: func(ctx *cli.Context) error {
				conf, err := loadConfig(ctx.Args().Slice(), ctx.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				lg := logger.New(currentAppVersion, conf.Log)

				st, err := storage.InitStore(ctx.Context, conf)
				if err != nil {
					return fmt.Errorf("storage init: %w", err)
				}
				defer func() {
					if storageCloseErr := st.Close(); storageCloseErr != nil {
						lg.Errorw("storage close error", "error", storageCloseErr)
					}
				}()

				srv, err := permission.New(conf, st.PSQLConn())
				if err != nil {
					return fmt.Errorf("init permission service failed: %w", err)
				}
				if !migrations.ConfirmActions(ctx.Context, "Are you sure?", false) {
					return nil
				}
				srv.ClearPolicy()
				return nil
			},
		}, // permission.clear
	}
}

func prepareUserPermissionCommands(currentAppVersion string) []*cli.Command {
	return []*cli.Command{
		{
			Name:        "add",
			Description: "add permission role to given user",
			Action: func(ctx *cli.Context) error {
				conf, err := loadConfig(ctx.Args().Slice(), ctx.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				lg := logger.New(currentAppVersion, conf.Log)

				st, err := storage.InitStore(ctx.Context, conf)
				if err != nil {
					return fmt.Errorf("storage init: %w", err)
				}
				defer func() {
					if storageCloseErr := st.Close(); storageCloseErr != nil {
						lg.Errorw("storage close error", "error", storageCloseErr)
					}
				}()

				srv, err := permission.New(conf, st.PSQLConn())
				if err != nil {
					return fmt.Errorf("init permission service failed: %w", err)
				}
				ok, err := srv.AddUserRole(ctx.String("id"), models.UserRole(ctx.String("role")))
				if err != nil {
					return fmt.Errorf("add user role failed: %w", err)
				}
				result := "OK"
				if !ok {
					result = "user already has role"
				}
				_, err = fmt.Fprintln(os.Stdout, result)
				return err
			},
		}, // permission.user.add
		{
			Name:        "delete",
			Description: "deletes permission role for given user",
			Action: func(ctx *cli.Context) error {
				conf, err := loadConfig(ctx.Args().Slice(), ctx.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				lg := logger.New(currentAppVersion, conf.Log)

				st, err := storage.InitStore(ctx.Context, conf)
				if err != nil {
					return fmt.Errorf("storage init: %w", err)
				}
				defer func() {
					if storageCloseErr := st.Close(); storageCloseErr != nil {
						lg.Errorw("storage close error", "error", storageCloseErr)
					}
				}()
				srv, err := permission.New(conf, st.PSQLConn())
				if err != nil {
					return fmt.Errorf("init permission service failed: %w", err)
				}
				ok, err := srv.DeleteUserRole(ctx.String("id"), models.UserRole(ctx.String("role")))
				if err != nil {
					return fmt.Errorf("delete user role failed: %w", err)
				}
				result := "OK"
				if !ok {
					result = "user did not have role"
				}
				_, err = fmt.Fprintln(os.Stdout, result)
				return err
			},
		}, // permission.user.delete
	}
}

func prepareTransactionsCommands(currentAppVersion string) []*cli.Command {
	return []*cli.Command{
		{
			Name: "restore",
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:    "blockchains",
					Aliases: []string{"b"},
					Usage:   "target blockchains to restore",
				},
			},
			Action: func(ctx *cli.Context) error {
				conf, err := loadConfig(ctx.Args().Slice(), ctx.StringSlice("configs"))
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}

				lg := logger.New(currentAppVersion, conf.Log)
				st, err := storage.InitStore(ctx.Context, conf)
				if err != nil {
					return fmt.Errorf("storage init: %w", err)
				}

				currService := currency.New(conf, st)
				exrateService, err := exrate.New(conf, currService, lg, st)
				if err != nil {
					return fmt.Errorf("init exrate service failed: %w", err)
				}

				currConvService := currconv.New(exrateService)
				eprCl, err := epr.NewClient(conf.EProxy.GRPC.Addr, epr.WithConnectrpcOpts(
					connect.WithGRPC(),
				))
				if err != nil {
					return fmt.Errorf("new epr client failed: %w", err)
				}

				eProxyService := eproxy.New(eprCl)
				eventListener := event.New()

				transactionService := transactions.New(lg, st, eProxyService, currConvService, eventListener, nil)

				blockchains := ctx.StringSlice("blockchains")

				return transactionService.RestoreAllWallets(ctx.Context, blockchains)
			},
		},
	}
}

func prepareConfigCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "genenvs",
			Usage: "generate markdown for all envs and config yaml template",
			Action: func(_ *cli.Context) error {
				conf := new(config.Config)

				envMarkdown, err := xconfig.GenerateMarkdown(conf, xconfig.WithEnvPrefix(envPrefix))
				if err != nil {
					return fmt.Errorf("failed to generate markdown: %w", err)
				}
				envMarkdown = fmt.Sprintf("# Environment variables\n\nAll envs have prefix `%s_`\n\n%s", envPrefix, envMarkdown)
				if err := os.WriteFile("ENVS.md", []byte(envMarkdown), 0o600); err != nil {
					return err
				}

				buf := bytes.NewBuffer(nil)
				enc := yaml.NewEncoder(buf, yaml.Indent(2))
				defer enc.Close()

				if err := enc.Encode(conf); err != nil {
					return fmt.Errorf("failed to encode yaml: %w", err)
				}

				if err := os.WriteFile("configs/config.template.yaml", buf.Bytes(), 0o600); err != nil {
					return fmt.Errorf("failed to write file: %w", err)
				}

				return nil
			},
		},
		{
			Name:  "validate",
			Usage: "validate config without starting the server",
			Flags: []cli.Flag{cfgPathsFlag()},
			Action: func(ctx *cli.Context) error {
				_, err := config.Load[config.Config](ctx.StringSlice("configs"), envPrefix)
				if err != nil {
					return err
				}
				return nil
			},
		},
	}
}

func cfgPathsFlag() *cli.StringSliceFlag {
	return &cli.StringSliceFlag{
		Name:    "configs",
		Aliases: []string{"c"},
		Value:   cli.NewStringSlice(defaultConfigPath),
		Usage:   "allows you to use your own paths to configuration files, separated by commas (config.yaml,config.prod.yml,.env)",
	}
}

func loadConfig(_, configPaths []string) (*config.Config, error) {
	conf, err := config.Load[config.Config](configPaths, envPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return conf, nil
}
