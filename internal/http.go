package kuura

import (
	"log/slog"
	"net/http"

	m "github.com/kymppi/kuura/internal/middleware"
)

func NewHTTPServer(logger *slog.Logger, config *Config) *http.Server {
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
