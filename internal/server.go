package kuura

import (
	"context"
	"fmt"
	"log/slog"
)

func RunServer(ctx context.Context, logger *slog.Logger, config *Config) error {
	dbManager := NewDatabaseManager(logger)

	logger.Info("Automatic migration apply mode is enabled")

	pool, err := dbManager.Connect(config.DATABASE_URL)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer pool.Close()

	if err := HandleMigrations(logger, dbManager, config.RUN_MIGRATIONS); err != nil {
		return err
	}

	server := newHTTPServer(logger, config)
	errChan := make(chan error, 1)
	go startHTTPServer(server, logger, errChan)

	return waitForShutdown(ctx, server, logger, errChan)
}
