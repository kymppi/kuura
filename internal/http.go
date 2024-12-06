package kuura

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	m "github.com/kymppi/kuura/internal/middleware"
)

func newHTTPServer(logger *slog.Logger, config *Config) *http.Server {
	mux := http.NewServeMux()

	addRoutes(
		mux,
		logger,
		config,
	)

	var handler http.Handler = mux

	handler = m.LoggingMiddleware(logger, handler)
	//TODO: opentelemetry tracing

	return &http.Server{
		Addr:    config.LISTEN,
		Handler: handler,
	}
}

func startHTTPServer(server *http.Server, logger *slog.Logger, errChan chan<- error) {
	logger.Info("Starting HTTP server", slog.String("addr", server.Addr))
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errChan <- fmt.Errorf("HTTP server error: %w", err)
	}
}
