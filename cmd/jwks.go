package cmd

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	kuura "github.com/kymppi/kuura/internal"
	"github.com/spf13/cobra"
)

func runJwks(logger *slog.Logger, config *kuura.Config) *cobra.Command {
	jwksCmd := &cobra.Command{
		Use:     "jwks",
		Aliases: []string{"jwk"},
		Short:   "Manage the JWKs of a specific service",
	}

	jwksCmd.AddCommand(jwkCreate(logger, config))

	return jwksCmd
}

func jwkCreate(logger *slog.Logger, config *kuura.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "create [service-id]",
		Short: "Create a new JWK for a service",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			serviceId, err := uuid.Parse(args[0])
			if err != nil {
				cmd.PrintErrf("Failed to parse serviceId: %s", err)
				return
			}

			queries, cleanup, err := kuura.InitializeDatabaseConnection(ctx, logger, config)
			if err != nil {
				cmd.PrintErrf("Failed to initialize database: %s", err)
				return
			}
			defer cleanup()

			jwkManager, err := kuura.InitializeJWKManager(ctx, logger, config, queries)
			if err != nil {
				cmd.PrintErrf("Failed to initialize jwk manager: %s", err)
				return
			}

			keyID, err := jwkManager.CreateKey(ctx, serviceId)
			if err != nil {
				cmd.PrintErrf("Failed to create JWK for service ID %s: %v\n", serviceId.String(), err)
				return
			}

			cmd.Printf("New key created with ID: %s for service %s\n", keyID, serviceId)
		},
	}
}
