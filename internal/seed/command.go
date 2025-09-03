package seed

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/pkg/database"
)

type Seeder struct {
	db      *database.PostgresClient
	baseDir string
	fs      embed.FS
}

func NewSeeder(ctx context.Context, conf *config.Config, fs embed.FS) (*Seeder, error) {
	db, err := database.NewPostgresClient(ctx, conf.Postgres)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	if err = db.EnsureSchemaMigrationsReady(ctx); err != nil {
		return nil, errors.New("database schema is dirty")
	}
	return &Seeder{
		db:      db,
		baseDir: conf.Seed.Base,
		fs:      fs,
	}, nil
}

func (s *Seeder) Run(ctx context.Context, opt *Options) (err error) {
	if len(opt.File) > 0 {
		return s.execSeed(ctx, opt.File)
	}

	if len(opt.Name) > 0 {
		return s.execSeed(ctx, fmt.Sprintf("%s.%s.sql", opt.Name, opt.Mode))
	}

	err = s.execAllSeeds(ctx, opt.Mode)
	log.Println("seed: done")
	return err
}

func (s *Seeder) execAllSeeds(ctx context.Context, mode Mode) error {
	log.Println("seed: scan database seeds in dir ", s.baseDir)
	entries, err := s.fs.ReadDir(s.baseDir)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}

	suffix := "." + string(mode) + ".sql"
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), suffix) {
			if fileExecErr := s.execSeed(ctx, entry.Name()); fileExecErr != nil {
				return fmt.Errorf("exec seed: %w", fileExecErr)
			}
		}
	}

	return nil
}

func (s *Seeder) execSeed(ctx context.Context, filePath string) (err error) {
	fmt.Printf("EXEC %s ...", filePath)
	var data []byte
	data, err = s.fs.ReadFile(s.baseDir + "/" + filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	_, err = s.db.DB.Exec(ctx, string(data))
	if err != nil {
		fmt.Println("[FAIL]")
	} else {
		fmt.Println("[OK]")
	}

	return err
}
