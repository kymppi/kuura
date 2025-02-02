package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"

	kuura "github.com/kymppi/kuura/internal"
	"github.com/kymppi/kuura/internal/srp"
	"github.com/kymppi/kuura/internal/users"
	"github.com/manifoldco/promptui"
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

			userService := users.NewUserService(logger, queries)

			prime, ok := new(big.Int).SetString(config.SRP_PRIME, 16)
			if !ok {
				cmd.PrintErrf("Invalid SRP_PRIME value, expected hex")
				return
			}

			srpKey, err := srp.GenerateRandomKey(prime)
			if err != nil {
				cmd.PrintErrf("Failed to create a random key for SRP: %s", err)
				return
			}

			options := &srp.SRPOptions{
				PrimeHex:  config.SRP_PRIME,
				Generator: config.SRP_GENERATOR,
			}

			srpClient, err := srp.NewSRPClient(options, srpKey)
			if err != nil {
				cmd.PrintErrf("Failed to create SRP struct: %s", err)
				return
			}

			srp_salt_int, srp_verifier_int, err := srpClient.Register(username, password)
			if err != nil {
				cmd.PrintErrf("SRP register failed: %s", err)
			}

			srp_salt := srp_salt_int.Text(16) // hex
			srp_verifier := srp_verifier_int.Text(16)

			uid, err := userService.Register(ctx, username, srp_salt, srp_verifier)
			if err != nil {
				cmd.PrintErrf("Failed to create user: %s", err)
				return
			}

			cmd.Printf("User '%s' created successfully with id %s!", username, uid)
		},
	}
}
