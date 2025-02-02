package kuura

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/kymppi/kuura/internal/jwks"
	"github.com/kymppi/kuura/internal/m2m"
	m "github.com/kymppi/kuura/internal/middleware"
	"github.com/kymppi/kuura/internal/srp"
)

func newHTTPServer(
	logger *slog.Logger,
	config *Config,
	jwkManager *jwks.JWKManager,
	m2mService *m2m.M2MService,
	srpOptions *srp.SRPOptions,
	frontendFS embed.FS,
) *http.Server {
	mux := http.NewServeMux()

	serverLogger := logger.With(slog.String("type", "main"))

	addMainRoutes(
		mux,
		serverLogger,
		jwkManager,
		m2mService,
		srpOptions,
		frontendFS,
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
	m2mService *m2m.M2MService,
) *http.Server {
	mux := http.NewServeMux()

	serverLogger := logger.With(slog.String("type", "management"))

	addManagementRoutes(
		mux,
		serverLogger,
		jwkManager,
		m2mService,
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
