package cmd

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	kuura "github.com/kymppi/kuura/internal"
	"github.com/kymppi/kuura/internal/m2m"
	"github.com/spf13/cobra"
)

func runM2M(logger *slog.Logger, config *kuura.Config) *cobra.Command {
	m2mCmd := &cobra.Command{
		Use:     "m2m",
		Aliases: []string{"machine2machine"},
		Short:   "Manage Machine-To-Machine related options",
	}

	m2mCmd.AddCommand(m2mRoleTemplateCreate(logger, config))
	m2mCmd.AddCommand(m2mRoleTemplateList(logger, config))

	return m2mCmd
}

func m2mRoleTemplateCreate(logger *slog.Logger, config *kuura.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "create [service-id] [template-name] [...roles]",
		Short: "Create a new template with roles",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			serviceId, err := uuid.Parse(args[0])
			if err != nil {
				cmd.PrintErrf("Failed to parse serviceId: %s", err)
				return
			}
			templateId := args[1]
			roles := args[2:] // Remaining arguments are roles

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

			m2mService := m2m.NewM2MService(queries, config.JWT_ISSUER, jwkManager)

			if err := m2mService.CreateRoleTemplate(ctx, serviceId, templateId, roles); err != nil {
				cmd.PrintErrf("Failed to create role template: %s", err)
				return
			}

			cmd.Printf("Role template '%s' created successfully with roles: %v\n", templateId, roles)
		},
	}
}

func m2mRoleTemplateList(logger *slog.Logger, config *kuura.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "list [service-id]",
		Short: "List all role templates",
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

			m2mService := m2m.NewM2MService(queries, config.JWT_ISSUER, jwkManager)

			templates, err := m2mService.GetRoleTemplates(ctx, serviceId)
			if err != nil {
				cmd.PrintErrf("Failed to list role templates: %s", err)
				return
			}

			if len(templates) == 0 {
				cmd.Println("No role templates found.")
				return
			}

			cmd.Println("Role Templates:")
			for _, template := range templates {
				cmd.Printf(" - %s: %v\n", template.Id, template.Roles)
			}
		},
	}
}
