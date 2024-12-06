package kuura

import (
	"log/slog"
	"net/http"
)

func addRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	config *Config,
) {
	mux.Handle("/", http.NotFoundHandler())
}
