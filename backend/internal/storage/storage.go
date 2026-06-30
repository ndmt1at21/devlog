// Package storage selects and constructs the domain.Store backend (memory or
// mysql) from configuration, mirroring the IAM service's backend factory.
package storage

import (
	"context"
	"fmt"

	"github.com/ndmt1at21/devlog/backend/internal/config"
	"github.com/ndmt1at21/devlog/backend/internal/domain"
	"github.com/ndmt1at21/devlog/backend/internal/repository/memory"
	"github.com/ndmt1at21/devlog/backend/internal/repository/mysql"
)

// New builds the store for the configured driver.
func New(ctx context.Context, cfg config.Config) (domain.Store, error) {
	switch cfg.DBDriver {
	case "mysql":
		return mysql.New(ctx, cfg.DBDSN)
	case "memory", "":
		return memory.New(), nil
	default:
		return nil, fmt.Errorf("unknown DB_DRIVER %q", cfg.DBDriver)
	}
}
