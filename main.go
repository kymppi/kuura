package main

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

	// Global flags
	debugMode bool
)

func main() {
	if GitSHA == "" {
		GitSHA = "unknown"
	}
	if Branch == "" {
		Branch = "unknown"
	}

	rootCmd = &cobra.Command{
		Use:     "kuura",
		Short:   "Kuura Authentication Server",
		Long:    `Kuura is an authentication server with great M2M support.`,
		Run:     kuura.RootCommand,
		Version: formatVersion(GitSHA, Branch),
	}

	rootCmd.PersistentFlags().BoolVar(
		&debugMode,
		"debug",
		false,
		"enable debug logging",
	)

	kuura.SetLoggerDebugMode(debugMode)

	logger := kuura.ProvideLogger()

	// rootCmd.AddCommand(kuura.VersionCommand())
	rootCmd.AddCommand(kuura.MigrateCommand(logger))

	if err := rootCmd.Execute(); err != nil {
		logger.Error("Command execution failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func formatVersion(sha, branch string) string {
	return fmt.Sprintf("%s (%s)", branch, sha)
}
