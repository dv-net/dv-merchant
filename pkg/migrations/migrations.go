package migrations

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dv-net/dv-merchant/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
)

type Migration struct {
	logger *logger
	config Config
}

func New(l Logger, conf Config) (*Migration, error) {
	if l == nil {
		return nil, fmt.Errorf("logger is required")
	}

	if err := conf.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &Migration{
		logger: newLogger(l),
		config: conf,
	}, nil
}

func (s *Migration) prepare(ctx context.Context) (*migrate.Migrate, database.Driver, error) {
	if !s.config.DisableConfirmation && !ConfirmActions(ctx, "Are you sure?", false) {
		return nil, nil, fmt.Errorf("user cancelled")
	}

	var (
		driver database.Driver
		dbName string
		err    error
	)

	if s.config.DBDriver == DBDriverMySQL {
		return nil, nil, fmt.Errorf("currently mysql is not supported")
	}

	conf, err := pgx.ParseConfig(s.config.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("parse postgres database dsn failed: %w", err)
	}
	dbName = conf.Database
	params := sql.PostgresMigrationParams()
	driver, err = s.postgresDriver(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("initialize postgres database connection: %w", err)
	}

	sourceDriver, err := iofs.New(params.EmbedFs, params.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("driver from embed.FS and a relative path [%s]: %w", params.Path, err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, dbName, driver)
	if err != nil {
		return nil, nil, err
	}
	m.Log = s.logger

	return m, driver, nil
}

// Up applies the database schema by the specified number of steps.
func (s *Migration) Up(ctx context.Context, steps int) error {
	m, instance, err := s.prepare(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = instance.Close() }()

	if steps > 0 {
		if err := m.Steps(steps); err != nil {
			return err
		}
	} else {
		if err := m.Up(); err != nil {
			return err
		}
	}

	return nil
}

// Down rolls back the database schema by the specified number of steps.
func (s *Migration) Down(ctx context.Context, steps int) error {
	m, instance, err := s.prepare(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = instance.Close() }()

	if steps == 0 {
		return fmt.Errorf("steps must be greater than 0")
	}

	if err := m.Steps(steps * -1); err != nil {
		return err
	}

	return nil
}

// Drop removes all database schema.
func (s *Migration) Drop(ctx context.Context) error {
	m, instance, err := s.prepare(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = instance.Close() }()

	if err := m.Drop(); err != nil {
		return err
	}

	return nil
}

// Version returns the current database schema version.
func (s *Migration) Version(ctx context.Context) (uint, bool, error) {
	m, instance, err := s.prepare(ctx)
	if err != nil {
		return 0, false, err
	}
	defer func() { _ = instance.Close() }()

	version, dirty, err := m.Version()
	if err != nil {
		return 0, false, err
	}

	return version, dirty, nil
}

func ConfirmActions(ctx context.Context, label string, priority bool) bool {
	choices := "Y/n"
	if !priority {
		choices = "y/N"
	}

	if _, err := fmt.Fprintf(os.Stderr, "%s (%s) ", label, choices); err != nil {
		panic(err)
	}

	reader := func(r io.Reader) <-chan string {
		lines := make(chan string)
		go func() {
			defer close(lines)
			scan := bufio.NewScanner(r)
			for scan.Scan() {
				s := scan.Text()
				lines <- s
			}
		}()
		return lines
	}

	for {
		select {
		case <-ctx.Done():
			return false
		case str, ok := <-reader(os.Stdin):
			if !ok {
				return false
			}
			s := strings.TrimSpace(str)
			if s == "" {
				return priority
			}
			s = strings.ToLower(s)
			if s == "y" || s == "yes" {
				return true
			}
			if s == "n" || s == "no" {
				return false
			}
		}
	}
}
