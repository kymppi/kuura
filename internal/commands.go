package kuura

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func RootCommand(cmd *cobra.Command, args []string) {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	logger := ProvideLogger()

	if err := runApplication(ctx, logger); err != nil {
		logger.Error("Fatal Application Error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func runApplication(ctx context.Context, logger *slog.Logger) error {
	logger.Info("Loading configuration...")
	config, err := ParseConfig()
	if err != nil {
		return fmt.Errorf("loading configuration failed: %w", err)
	}
	logger.Info("Configuration loaded successfully")

	dbManager := NewDatabaseManager(logger)

	logger.Info("Automatic migration apply mode is enabled")

	pool, err := dbManager.Connect(config.DATABASE_URL)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer pool.Close()

	if err := handleMigrations(dbManager, config.RUN_MIGRATIONS); err != nil {
		return err
	}

	server := newHTTPServer(logger, config)
	errChan := make(chan error, 1)
	go startHTTPServer(server, logger, errChan)

	return waitForShutdown(ctx, server, logger, errChan)
}

func VersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of the application",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Kuura Application Server")
			fmt.Println("Version: v0.1.0")
		},
	}
}

func MigrateCommand(logger *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := ParseConfig()
			if err != nil {
				logger.Error("Failed to load configuration", slog.String("error", err.Error()))
				return
			}

			dbManager := NewDatabaseManager(logger)

			pool, err := dbManager.Connect(config.DATABASE_URL)
			if err != nil {
				logger.Error("Failed to connect to database", slog.String("error", err.Error()))
				return
			}
			defer pool.Close()

			if err := handleMigrations(dbManager, true); err != nil {
				logger.Error("Error while applying migrations", slog.String("error", err.Error()))
			}
		},
	}
}
