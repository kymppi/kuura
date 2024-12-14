package kuura

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/kymppi/kuura/internal/db_gen"
	"github.com/kymppi/kuura/internal/jwks"
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

func InitializeJWKManager(ctx context.Context, logger *slog.Logger, config *Config, queries *db_gen.Queries) (*jwks.JWKManager, error) {
	encryptionKey, err := loadEncryptionKey(config.JWK_KEK_PATH)
	if err != nil {
		logger.Error("Failed to load encryption key", slog.String("error", err.Error()))
		return nil, err
	}

	storage := jwks.NewPostgresQLKeyStorage(queries, encryptionKey)

	return jwks.NewJWKManager(storage), nil
}

func loadEncryptionKey(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read encryption key file: %w", err)
	}
	return data, nil
}
