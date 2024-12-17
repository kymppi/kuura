package kuura

import (
	"context"
	"log/slog"
	"net/http"
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

	mainServer := newHTTPServer(logger, config, jwkManager)
	managementServer := newManagementServer(logger, config, jwkManager)

	errChan := make(chan error, 2)

	go startHTTPServer(mainServer, logger, errChan, "main")
	go startHTTPServer(managementServer, logger, errChan, "management")

	return waitForShutdown(ctx, []*http.Server{mainServer, managementServer}, logger, errChan)
}
