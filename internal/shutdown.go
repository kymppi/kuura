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
	servers []*http.Server,
	logger *slog.Logger,
	errChan <-chan error,
) error {
	combinedErrChan := make(chan error, len(servers)+1) // Buffer for all servers + ctx.Done()

	// context cancellation
	go func() {
		<-ctx.Done()
		for _, server := range servers {
			if err := performGracefulShutdown(ctx, server, logger); err != nil {
				combinedErrChan <- err
			}
		}
		close(combinedErrChan)
	}()

	// errors from all servers
	go func() {
		for err := range errChan {
			combinedErrChan <- err
		}
		close(combinedErrChan)
	}()

	// 1st occurred error
	for err := range combinedErrChan {
		if err != nil {
			return err
		}
	}

	return nil
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
