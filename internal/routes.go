package kuura

import (
	"log/slog"
	"net/http"

	"github.com/kymppi/kuura/internal/endpoints"
	"github.com/kymppi/kuura/internal/jwks"
)

func addRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	jwkManager *jwks.JWKManager,
) {
	mux.Handle("/", http.NotFoundHandler())
	mux.Handle("GET /v1/{serviceId}/jwks.json", endpoints.V1JwksHandler(logger, jwkManager))
}
