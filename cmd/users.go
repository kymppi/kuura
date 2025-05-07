package cmd

import (
	"context"
	"crypto"
	"fmt"
	"log/slog"

	kuura "github.com/kymppi/kuura/internal"
	"github.com/kymppi/kuura/internal/services"
	"github.com/kymppi/kuura/internal/settings"
	"github.com/manifoldco/promptui"
	"github.com/opencoff/go-srp"
	"github.com/spf13/cobra"
)

func runUsers(logger *slog.Logger, config *kuura.Config) *cobra.Command {
	usersCmd := &cobra.Command{
		Use:     "user",
		Aliases: []string{"users"},
		Short:   "Manage users",
	}

	usersCmd.AddCommand(usersCreate(logger, config))

	return usersCmd
}

func usersCreate(logger *slog.Logger, config *kuura.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "create [username]",
		Short: "Create a new user",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			username := args[0]
			prompt := promptui.Prompt{
				Label: "Password",
				Mask:  '*',
				Validate: func(s string) error {
					if len(s) < 16 {
						return fmt.Errorf("password must be at least 16 characters long")
					}

					return nil
				},
			}

			password, err := prompt.Run()
			if err != nil {
				cmd.PrintErrf("Error reading password: %s", err)
				return
			}

			queries, cleanup, err := kuura.InitializeDatabaseConnection(ctx, logger, config)
			if err != nil {
				cmd.PrintErrf("Failed to initialize database: %s", err)
				return
			}
			defer cleanup()

			settingsService := settings.NewSettingsService(logger, queries)
			serviceManager := services.NewServiceManager(logger, queries, settingsService)
			jwkManager, err := kuura.InitializeJWKManager(ctx, logger, config, queries)
			if err != nil {
				cmd.PrintErrf("Failed to initialize jwk manager: %s", err)
				return
			}

			vh, err := generateVerifierHash(username, password)
			if err != nil {
				cmd.PrintErrf("Failed to generate a verifier hash: %s", err)
				return
			}

			userService, err := kuura.InitializeUserService(ctx, logger, config, queries, jwkManager, serviceManager)
			if err != nil {
				panic(err)
			}

			uid, err := userService.Register(ctx, username, vh)
			if err != nil {
				cmd.PrintErrf("Failed to create user: %s", err)
				return
			}

			cmd.Printf("User '%s' created successfully with id %s!", username, uid)
		},
	}
}

func generateVerifierHash(username, password string) (string, error) {
	s, err := srp.NewWithHash(crypto.SHA256, 4096)
	if err != nil {
		return "", fmt.Errorf("failed to create SRP instance: %w", err)
	}

	v, err := s.Verifier([]byte(username), []byte(password))
	if err != nil {
		return "", fmt.Errorf("failed to create verifier: %w", err)
	}

	_, vh := v.Encode()
	return vh, nil
}
