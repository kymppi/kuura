package kuura

import (
	"embed"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/kymppi/kuura/internal/constants"
	"github.com/kymppi/kuura/internal/endpoints"
	"github.com/kymppi/kuura/internal/jwks"
	"github.com/kymppi/kuura/internal/m2m"
	"github.com/kymppi/kuura/internal/services"
	"github.com/kymppi/kuura/internal/users"
)

func addMainRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	jwkManager *jwks.JWKManager,
	m2mService *m2m.M2MService,
	frontendFS embed.FS,
	userService *users.UserService,
	serviceManager *services.ServiceManager,
	publicKuuraDomain string,
	jwtIssuer string,
) {
	mux.Handle("GET /v1/service/{serviceId}/jwks.json", endpoints.V1JwksHandler(logger, jwkManager))
	mux.Handle("GET /v1/service/{serviceId}", endpoints.V1_ServiceInfo(logger, serviceManager))

	mux.Handle("POST /v1/m2m/access", endpoints.V1M2MRefreshAccessToken(logger, m2mService))

	mux.Handle("POST /v1/logout", endpoints.V1_User_Logout(logger, userService, publicKuuraDomain, jwkManager, jwtIssuer))
	mux.Handle("POST /v1/user/tokens/external", endpoints.V1_User_ExternalTokens(logger, userService))
	mux.Handle("POST /v1/user/login/external", endpoints.V1_User_LoginExternal(logger, userService, jwkManager, jwtIssuer))
	mux.Handle(fmt.Sprintf("POST %s", constants.INTERNAL_USER_REFRESH_PATH), endpoints.V1_User_RefreshInternalToken(logger, userService, publicKuuraDomain))

	mux.Handle("GET /v1/me", endpoints.V1_ME(logger, userService, jwkManager, jwtIssuer))

	mux.Handle("POST /v1/srp/begin", endpoints.V1_SRP_ClientBegin(logger, userService))
	mux.Handle("POST /v1/srp/verify", endpoints.V1_SRP_ClientVerify(logger, userService, publicKuuraDomain))

	mux.Handle("GET /", endpoints.FrontendHandler(logger, frontendFS))

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
