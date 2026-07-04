// Command seed loads the design's demo content (sample articles, series and
// comments) into a MySQL database. It is a separate, explicit operator action —
// the server itself never seeds — so production databases stay clean unless
// someone deliberately runs this against them.
//
// It reads the same environment configuration as the server (DB_DRIVER=mysql
// and DB_DSN are required) and applies schema migrations before upserting the
// seed rows, so it is safe to run against a fresh database. Seeding is
// idempotent: articles/series are upserted by slug and demo comments are only
// inserted when the comments table is empty.
//
//	DB_DRIVER=mysql DB_DSN='…' go run ./cmd/seed
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ndmt1at21/devlog/backend/internal/config"
	"github.com/ndmt1at21/devlog/backend/internal/platform/logger"
	"github.com/ndmt1at21/devlog/backend/internal/repository/mysql"
)

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", "err", err.Error())
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(os.Stdout, logger.Options{Level: cfg.LogLevel, Format: cfg.LogFormat})
	slog.SetDefault(log)

	// Seeding only makes sense for a persistent store. The memory driver already
	// boots pre-seeded and is thrown away on exit, so there is nothing to do.
	if cfg.DBDriver != "mysql" {
		return fmt.Errorf("seed requires DB_DRIVER=mysql (got %q); the memory store is always pre-seeded", cfg.DBDriver)
	}

	ctx := context.Background()
	store, err := mysql.New(ctx, cfg.DBDSN)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Seed(ctx); err != nil {
		return fmt.Errorf("seed: %w", err)
	}
	log.Info("seed complete: demo content loaded into mysql")
	return nil
}
