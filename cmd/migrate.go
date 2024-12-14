package cmd

import (
	"log/slog"

	kuura "github.com/kymppi/kuura/internal"
	"github.com/spf13/cobra"
)

func runMigrate(logger *slog.Logger, config *kuura.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Run: func(cmd *cobra.Command, args []string) {
			dbManager := kuura.NewDatabaseManager(logger)

			pool, err := dbManager.Connect(config.DATABASE_URL)
			if err != nil {
				logger.Error("Failed to connect to database", slog.String("error", err.Error()))
				return
			}
			defer pool.Close()

			if err := kuura.HandleMigrations(logger, dbManager, true); err != nil {
				logger.Error("Error while applying migrations", slog.String("error", err.Error()))
			}
		},
	}
}
