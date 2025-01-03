package kuura

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/kymppi/kuura/internal/m2m"
)

func RunServer(ctx context.Context, logger *slog.Logger, config *Config) error {
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

	mainServer := newHTTPServer(logger, config, jwkManager, m2mService)
	managementServer := newManagementServer(logger, config, jwkManager, m2mService)

	errChan := make(chan error, 2)

	go startHTTPServer(mainServer, logger, errChan, "main")
	go startHTTPServer(managementServer, logger, errChan, "management")

	return waitForShutdown(ctx, []*http.Server{mainServer, managementServer}, logger, errChan)
}
