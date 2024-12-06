package kuura

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func waitForShutdown(
	ctx context.Context,
	server *http.Server,
	logger *slog.Logger,
	errChan <-chan error,
) error {
	combinedErrChan := make(chan error, 1)

	go func() {
		select {
		case <-ctx.Done():
			combinedErrChan <- performGracefulShutdown(ctx, server, logger)
		case err := <-errChan:
			combinedErrChan <- err
		}
	}()

	return <-combinedErrChan
}

func performGracefulShutdown(
	ctx context.Context,
	server *http.Server,
	logger *slog.Logger,
) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	logger.Info("Initiating graceful shutdown")

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Shutdown failed", slog.String("error", err.Error()))
		return fmt.Errorf("server shutdown error: %w", err)
	}

	logger.Info("Server shutdown completed successfully")
	return nil
}
