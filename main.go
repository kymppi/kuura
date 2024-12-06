package main

import (
	"log/slog"
	"os"

	kuura "github.com/kymppi/kuura/internal"
	"github.com/spf13/cobra"
)

var (
	rootCmd *cobra.Command
	logger  *slog.Logger

	// Global flags
	debugMode bool
)

func main() {
	rootCmd = &cobra.Command{
		Use:   "kuura",
		Short: "Kuura Authentication Server",
		Long:  `Kuura is an authentication server with great M2M support.`,
		Run:   kuura.RootCommand,
	}

	rootCmd.PersistentFlags().BoolVar(
		&debugMode,
		"debug",
		false,
		"enable debug logging",
	)

	kuura.SetLoggerDebugMode(debugMode)

	logger := kuura.ProvideLogger()

	rootCmd.AddCommand(kuura.VersionCommand())
	rootCmd.AddCommand(kuura.MigrateCommand(logger))

	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command execution failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
