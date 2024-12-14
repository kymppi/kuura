package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	kuura "github.com/kymppi/kuura/internal"
	"github.com/kymppi/kuura/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func runServices(logger *slog.Logger, config *kuura.Config) *cobra.Command {
	servicesCmd := &cobra.Command{
		Use:   "services",
		Short: "Manage application services",
		Long:  `Interact with and manage application services in the Kuura authentication system.`,
	}

	servicesCmd.AddCommand(serviceList(logger, config))
	servicesCmd.AddCommand(serviceCreate(logger, config))
	servicesCmd.AddCommand(serviceDelete(logger, config))

	return servicesCmd
}

func serviceCreate(logger *slog.Logger, config *kuura.Config) *cobra.Command {
	var (
		name     string
		audience string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new service",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			queries, cleanup, err := kuura.InitializeDatabaseConnection(ctx, logger, config)
			if err != nil {
				logger.Error("Failed to initialize database", slog.String("error", err.Error()))
				return
			}
			defer cleanup()

			serviceManager := kuura.NewServiceManager(queries)

			id, err := serviceManager.CreateService(ctx, name, audience)
			if err != nil {
				logger.Error("Failed to create service", slog.String("error", err.Error()))
				return
			}

			cmd.Println("Service created successfully:")
			cmd.Printf("ID: %v\n", id)
			cmd.Printf("Name: %s\n", name)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Name of the service")
	cmd.Flags().StringVarP(&audience, "audience", "a", "", "JWT audience for the service")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("audience")

	return cmd
}

func serviceDelete(logger *slog.Logger, config *kuura.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete a service by ID",
		Args:  cobra.ExactArgs(1), // serviceId
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			serviceId, err := uuid.Parse(args[0])
			if err != nil {
				cmd.PrintErrf("Failed to parse UUIDv7: %v\n", err)
				os.Exit(1)
			}

			queries, cleanup, err := kuura.InitializeDatabaseConnection(ctx, logger, config)
			if err != nil {
				logger.Error("Failed to initialize database", slog.String("error", err.Error()))
				return
			}
			defer cleanup()

			serviceManager := kuura.NewServiceManager(queries)

			err = serviceManager.DeleteService(ctx, serviceId)
			if err != nil {
				logger.Error("Failed to delete service", slog.String("error", err.Error()))
				return
			}

			cmd.Printf("Service %s deleted successfully\n", serviceId)
		},
	}
}

func serviceList(logger *slog.Logger, config *kuura.Config) *cobra.Command {
	var (
		outputFormat string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all services",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			queries, cleanup, err := kuura.InitializeDatabaseConnection(ctx, logger, config)
			if err != nil {
				logger.Error("Failed to initialize database", slog.String("error", err.Error()))
				return
			}
			defer cleanup()

			serviceManager := kuura.NewServiceManager(queries)

			services, err := serviceManager.GetServices(ctx)
			if err != nil {
				logger.Error("Failed to list services", slog.String("error", err.Error()))
				return
			}

			switch outputFormat {
			case "json":
				if err := outputServicesJSON(services, cmd.OutOrStdout()); err != nil {
					logger.Error("Failed to output JSON", slog.String("error", err.Error()))
				}
			case "yaml":
				if err := outputServicesYAML(services, cmd.OutOrStdout()); err != nil {
					logger.Error("Failed to output YAML", slog.String("error", err.Error()))
				}
			case "table":
				outputServicesTable(services, cmd.OutOrStdout())
			default:
				outputServicesRich(services, cmd.OutOrStdout())
			}
		},
	}

	// Add output format flag
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "rich",
		"Output format. Options: rich, table, json, yaml")

	return cmd
}

func outputServicesTable(services []*models.AppService, w io.Writer) {
	writer := tabwriter.NewWriter(w, 0, 8, 2, ' ', 0)
	defer writer.Flush()

	// Header
	fmt.Fprintln(writer, "ID\tNAME\tDESCRIPTION\tAUDIENCE\tCREATED AT\tMODIFIED AT")

	for _, service := range services {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			service.Id,
			service.Name,
			service.Description,
			service.JWTAudience,
			service.CreatedAt.Format(time.RFC3339),
			service.ModifiedAt.Format(time.RFC3339),
		)
	}
}

func outputServicesJSON(services []*models.AppService, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(services)
}

func outputServicesYAML(services []*models.AppService, w io.Writer) error {
	return yaml.NewEncoder(w).Encode(services)
}

func outputServicesRich(services []*models.AppService, w io.Writer) {
	for _, service := range services {
		fmt.Fprintf(w, "  Service: %s\n", service.Name)
		fmt.Fprintf(w, "  ID:          %s\n", service.Id)
		fmt.Fprintf(w, "  Audience:    %s\n", service.JWTAudience)
		fmt.Fprintf(w, "  Description: %s\n", service.Description)
		fmt.Fprintf(w, "  Created:     %s\n\n", service.CreatedAt.Format(time.RFC3339))
	}
}
