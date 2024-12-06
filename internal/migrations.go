package kuura

import (
	"errors"
	"fmt"

	"github.com/kymppi/kuura/internal/db_migrations"
)

func handleMigrations(dbManager *DatabaseManager, run bool) error {
	migrationSource := db_migrations.Migrations
	needsMigration, err := dbManager.CheckMigrations(migrationSource)

	if err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if needsMigration && !run {
		return errors.New("pending database migrations exist - run migrations before starting the application")
	}

	if needsMigration {
		logger.Info("Applying database migrations")
		if err := dbManager.ApplyMigrations(migrationSource); err != nil {
			return fmt.Errorf("failed to apply migrations: %w", err)
		}
		logger.Info("Migrations have been applied successfully")
	} else {
		logger.Info("Database schema up to date!")
	}

	return nil
}
