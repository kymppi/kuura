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

func RootCommand(config *Config, logger *slog.Logger) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(context.Background(),
			os.Interrupt,
			syscall.SIGTERM,
		)
		defer stop()

		if err := runApplication(ctx, logger, config); err != nil {
			logger.Error("Fatal Application Error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}
}

func runApplication(ctx context.Context, logger *slog.Logger, config *Config) error {
	dbManager := NewDatabaseManager(logger)

	logger.Info("Automatic migration apply mode is enabled")

	pool, err := dbManager.Connect(config.DATABASE_URL)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	defer pool.Close()

	if err := handleMigrations(logger, dbManager, config.RUN_MIGRATIONS); err != nil {
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

func MigrateCommand(logger *slog.Logger, config *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Run: func(cmd *cobra.Command, args []string) {
			dbManager := NewDatabaseManager(logger)

			pool, err := dbManager.Connect(config.DATABASE_URL)
			if err != nil {
				logger.Error("Failed to connect to database", slog.String("error", err.Error()))
				return
			}
			defer pool.Close()

			if err := handleMigrations(logger, dbManager, true); err != nil {
				logger.Error("Error while applying migrations", slog.String("error", err.Error()))
			}
		},
	}
}
