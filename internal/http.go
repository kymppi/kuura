package kuura

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/kymppi/kuura/internal/jwks"
	"github.com/kymppi/kuura/internal/m2m"
	m "github.com/kymppi/kuura/internal/middleware"
	"github.com/kymppi/kuura/internal/srp"
)

//go:embed templates/*.tmpl
var templates embed.FS

//go:embed static/*
var staticAssets embed.FS

func newHTTPServer(
	logger *slog.Logger,
	config *Config,
	jwkManager *jwks.JWKManager,
	m2mService *m2m.M2MService,
	srpOptions *srp.SRPOptions,
) *http.Server {
	mux := http.NewServeMux()

	serverLogger := logger.With(slog.String("type", "main"))

	tmpl := template.Must(template.ParseFS(templates, "templates/*.tmpl"))

	addMainRoutes(
		mux,
		serverLogger,
		jwkManager,
		m2mService,
		srpOptions,
		staticAssets,
		tmpl,
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
