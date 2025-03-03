package kuura

import (
	"context"
	"embed"
	"log/slog"
	"net/http"

	"github.com/kymppi/kuura/internal/m2m"
	"github.com/kymppi/kuura/internal/users"
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

	m2mService := m2m.NewM2MService(queries, config.JWT_ISSUER, jwkManager)
	userService := users.NewUserService(logger, queries, config.JWT_ISSUER, jwkManager)

	mainServer := newHTTPServer(logger, config, jwkManager, m2mService, frontendFS, userService)
	managementServer := newManagementServer(logger, config, jwkManager, m2mService)

	errChan := make(chan error, 2)

	go startHTTPServer(mainServer, logger, errChan, "main")
	go startHTTPServer(managementServer, logger, errChan, "management")

	return waitForShutdown(ctx, []*http.Server{mainServer, managementServer}, logger, errChan)
}
