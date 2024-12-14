package kuura

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kymppi/kuura/internal/db_gen"
)

// setup db, check migration status
func InitializeDatabaseConnection(ctx context.Context, logger *slog.Logger, config *Config) (*db_gen.Queries, func(), error) {
	dbManager := NewDatabaseManager(logger)

	pool, err := dbManager.Connect(config.DATABASE_URL)
	if err != nil {
		return nil, nil, fmt.Errorf("database connection failed: %w", err)
	}

	if err := HandleMigrations(logger, dbManager, config.RUN_MIGRATIONS); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("migrations failed: %w", err)
	}

	cleanup := func() {
		pool.Close()
		logger.Info("Database connection closed")
	}

	return db_gen.New(pool), cleanup, nil
}
