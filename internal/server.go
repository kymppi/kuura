package kuura

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/kymppi/kuura/internal/m2m"
	"github.com/kymppi/kuura/internal/services"
	"github.com/kymppi/kuura/internal/settings"
)

func RunServer(ctx context.Context, logger *slog.Logger, config *Config, frontendFS embed.FS) error {
	queries, cleanup, err := InitializeDatabaseConnection(ctx, logger, config)
	if err != nil {
		return err
	}
	defer cleanup()

	jwkManager, err := InitializeJWKManager(ctx, logger, config, queries)
	if err != nil {
		return err
	}

	settingsService := settings.NewSettingsService(logger, queries)
	serviceManager := services.NewServiceManager(logger, queries, settingsService)

	if err := serviceManager.CreateInternalServiceIfNotExists(ctx, config.PUBLIC_KUURA_DOMAIN); err != nil {
		return fmt.Errorf("faied to create internal service for kuura: %w", err)
	}

	m2mService := m2m.NewM2MService(queries, config.JWT_ISSUER, jwkManager)

	userService, err := InitializeUserService(ctx, logger, config, queries, jwkManager, serviceManager)
	if err != nil {
		return err
	}

	mainServer := newHTTPServer(logger, config, jwkManager, m2mService, frontendFS, userService, serviceManager)
	managementServer := newManagementServer(logger, config, jwkManager, m2mService)

	errChan := make(chan error, 2)

	go startHTTPServer(mainServer, logger, errChan, "main")
	go startHTTPServer(managementServer, logger, errChan, "management")

	return waitForShutdown(ctx, []*http.Server{mainServer, managementServer}, logger, errChan)
}
