package main

//go:generate sqlc generate

import (
	"fmt"
	"log/slog"
	"os"

	kuura "github.com/kymppi/kuura/internal"
	"github.com/spf13/cobra"
)

var (
	rootCmd *cobra.Command

	// Info
	GitSHA string
	Branch string
)

func main() {
	if GitSHA == "" {
		GitSHA = "unknown"
	}
	if Branch == "" {
		Branch = "unknown"
	}

	config, err := kuura.ParseConfig()
	if err != nil {
		fmt.Println("Failed to load configuration:", err)
		os.Exit(1)
	}

	loggerConfig := kuura.LoggerConfig{
		DebugEnabled:  config.DEBUG,
		PrettyEnabled: config.GO_ENV != "production",
	}
	loggerManager := kuura.NewLogger(loggerConfig)
	logger := loggerManager.Get()

	rootCmd = &cobra.Command{
		Use:     "kuura",
		Short:   "Kuura Authentication Server",
		Long:    `Kuura is an authentication server with great M2M support.`,
		Run:     kuura.RootCommand(config, logger),
		Version: formatVersion(GitSHA, Branch),
	}

	rootCmd.AddCommand(kuura.MigrateCommand(logger, config))

	logger.Info("Starting Kuura", slog.String("version", rootCmd.Version))

	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command execution failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func formatVersion(sha, branch string) string {
	return fmt.Sprintf("%s (%s)", branch, sha)
}
