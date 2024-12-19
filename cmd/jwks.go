package cmd

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"
	"sort"

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
	jwksCmd.AddCommand(jwkExport(logger, config))
	jwksCmd.AddCommand(jwkStatus(logger, config))

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

func jwkExport(logger *slog.Logger, config *kuura.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "export [service-id] [key-id]",
		Short: "Export a private key in PKCS #8",
		Args:  cobra.ExactArgs(2),
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

			privateJWK, err := jwkManager.Export(ctx, serviceId, args[1])
			if err != nil {
				cmd.PrintErrf("Failed to retrieve private key for service ID %s: %v\n", serviceId.String(), err)
				return
			}

			var privateKey ecdsa.PrivateKey
			err = privateJWK.Raw(&privateKey)
			if err != nil {
				cmd.PrintErrf("Failed to get private key from JWK: %v\n", err)
				return
			}

			err = exportPKCS8PrivateKeyToStdout(&privateKey)
			if err != nil {
				cmd.PrintErrf("Failed to export private key to PEM: %v", err)
				return
			}
		},
	}
}

func exportPKCS8PrivateKeyToStdout(privateKey *ecdsa.PrivateKey) error {
	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key into PKCS#8 format: %v", err)
	}

	err = pem.Encode(os.Stdout, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})
	if err != nil {
		return fmt.Errorf("failed to encode private key in PEM format: %v", err)
	}

	return nil
}

func jwkStatus(logger *slog.Logger, config *kuura.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "status [service-id]",
		Short: "View the key order, which key is currently used and which one is next.",
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

			keyStatuses, err := jwkManager.KeyStatus(ctx, serviceId)
			if err != nil {
				cmd.PrintErrf("Failed to get current key status: %s", err)
				return
			}

			sortedKeys := make([]string, 0, len(keyStatuses))
			keyOrder := map[string]int{"future": 0, "current": 1, "retired": 2}

			keySlice := make([]struct {
				ID     string
				Status string
			}, 0, len(keyStatuses))
			for keyID, status := range keyStatuses {
				keySlice = append(keySlice, struct {
					ID     string
					Status string
				}{keyID, status})
			}

			sort.Slice(keySlice, func(i, j int) bool {
				return keyOrder[keySlice[i].Status] < keyOrder[keySlice[j].Status]
			})

			for _, key := range keySlice {
				sortedKeys = append(sortedKeys, key.ID)
			}

			for _, keyID := range sortedKeys {
				fmt.Printf("- %s (%s)\n", keyID, keyStatuses[keyID])
			}
		},
	}
}
