package kuura

import (
	"context"
	"log/slog"
)

func RunServer(ctx context.Context, logger *slog.Logger, config *Config) error {
	_, cleanup, err := InitializeDatabaseConnection(ctx, logger, config)
	if err != nil {
		return err
	}
	defer cleanup()

	server := newHTTPServer(logger, config)
	errChan := make(chan error, 1)
	go startHTTPServer(server, logger, errChan)
	return waitForShutdown(ctx, server, logger, errChan)
}
