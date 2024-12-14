package cmd

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	kuura "github.com/kymppi/kuura/internal"
	"github.com/kymppi/kuura/internal/utils"
	"github.com/spf13/cobra"
)

var (
	GitSHA string
	Branch string

	rootCmd *cobra.Command
)

func NewRootCommand(config *kuura.Config, logger *slog.Logger) *cobra.Command {
	rootCmd = &cobra.Command{
		Use:     "kuura",
		Short:   "Kuura Authentication Server",
		Long:    `Kuura is an authentication server with great M2M support.`,
		Run:     runRoot(config, logger),
		Version: utils.FormatVersion(GitSHA, Branch),
	}

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if cmd.Parent() != nil {
			*logger = *slog.New(slog.NewJSONHandler(io.Discard, nil))
		}
	}

	// Subcommands
	rootCmd.AddCommand(runMigrate(logger, config))
	rootCmd.AddCommand(runServices(logger, config))
	rootCmd.AddCommand(runJwks(logger, config))

	return rootCmd
}

func runRoot(config *kuura.Config, logger *slog.Logger) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(context.Background(),
			os.Interrupt,
			syscall.SIGTERM,
		)
		defer stop()

		if err := kuura.RunServer(ctx, logger, config); err != nil {
			logger.Error("Fatal Application Error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}
}
