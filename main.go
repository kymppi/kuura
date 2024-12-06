package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	kuura "github.com/kymppi/kuura/internal"
)

func main() {
	logger := setupLogger()

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := runApplication(ctx, logger); err != nil {
		logger.Error("Application failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func setupLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func runApplication(ctx context.Context, logger *slog.Logger) error {
	logger.Info("Loading configuration...")
	config, err := loadConfiguration(logger)
	if err != nil {
		return fmt.Errorf("configuration load failed: %w", err)
	}

	server := kuura.NewHTTPServer(logger, config)

	errChan := make(chan error, 1)
	go startHTTPServer(server, logger, errChan)

	return waitForShutdown(ctx, server, logger, errChan)
}

func loadConfiguration(logger *slog.Logger) (*kuura.Config, error) {
	config, err := kuura.ParseConfig()
	if err != nil {
		logger.Error("Failed to load configuration", slog.String("error", err.Error()))
		return nil, err
	}
	logger.Info("Configuration loaded successfully")
	return config, nil
}

func startHTTPServer(server *http.Server, logger *slog.Logger, errChan chan<- error) {
	logger.Info("Starting HTTP server", slog.String("addr", server.Addr))
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errChan <- fmt.Errorf("HTTP server error: %w", err)
	}
}

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
