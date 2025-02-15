package kuura

import (
	"embed"
	"log/slog"
	"net/http"

	"github.com/kymppi/kuura/internal/endpoints"
	"github.com/kymppi/kuura/internal/jwks"
	"github.com/kymppi/kuura/internal/m2m"
	"github.com/kymppi/kuura/internal/users"
)

func addMainRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	jwkManager *jwks.JWKManager,
	m2mService *m2m.M2MService,
	frontendFS embed.FS,
	userService *users.UserService,
) {
	mux.Handle("/", http.NotFoundHandler())
	mux.Handle("GET /v1/{serviceId}/jwks.json", endpoints.V1JwksHandler(logger, jwkManager))
	mux.Handle("POST /v1/m2m/access", endpoints.V1M2MRefreshAccessToken(logger, m2mService))

	mux.Handle("POST /v1/srp/begin", endpoints.V1_SRP_ClientBegin(logger, userService))
	mux.Handle("POST /v1/srp/verify", endpoints.V1_SRP_ClientVerify(logger, userService))

	mux.Handle("GET /", endpoints.AstroHandler(logger, frontendFS))

	// authenticated management endpoints
}

func addManagementRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	jwkManager *jwks.JWKManager,
	m2mService *m2m.M2MService,
) {
	mux.Handle("/", http.NotFoundHandler())
	mux.Handle("GET /v1/{serviceId}/jwks.json", endpoints.V1JwksHandler(logger, jwkManager))

	// unauthenticated management endpoints
	mux.Handle("POST /v1/m2m/sessions", endpoints.V1CreateM2MSession(logger, m2mService))
}
