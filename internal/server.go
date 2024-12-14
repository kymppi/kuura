package kuura

import (
	"context"
	"log/slog"
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

	server := newHTTPServer(logger, config, jwkManager)
	errChan := make(chan error, 1)
	go startHTTPServer(server, logger, errChan)
	return waitForShutdown(ctx, server, logger, errChan)
}
