package kuura

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/kymppi/kuura/internal/jwks"
	m "github.com/kymppi/kuura/internal/middleware"
)

func newHTTPServer(
	logger *slog.Logger,
	config *Config,
	jwkManager *jwks.JWKManager,
) *http.Server {
	mux := http.NewServeMux()

	serverLogger := logger.With(slog.String("type", "main"))

	addMainRoutes(
		mux,
		serverLogger,
		jwkManager,
	)

	var handler http.Handler = mux

	handler = m.LoggingMiddleware(serverLogger, handler)
	//TODO: opentelemetry tracing

	return &http.Server{
		Addr:    config.LISTEN,
		Handler: handler,
	}
}

func newManagementServer(
	logger *slog.Logger,
	config *Config,
	jwkManager *jwks.JWKManager,
) *http.Server {
	mux := http.NewServeMux()

	serverLogger := logger.With(slog.String("type", "management"))

	addManagementRoutes(
		mux,
		serverLogger,
		jwkManager,
	)

	var handler http.Handler = mux

	handler = m.LoggingMiddleware(serverLogger, handler)
	//TODO: opentelemetry tracing

	return &http.Server{
		Addr:    config.MANAGEMENT_LISTEN,
		Handler: handler,
	}
}

func startHTTPServer(server *http.Server, logger *slog.Logger, errChan chan<- error, label string) {
	logger.Info("Starting HTTP server", slog.String("addr", server.Addr), slog.String("type", label))
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errChan <- fmt.Errorf("HTTP server (%s) error: %w", label, err)
	}
}
