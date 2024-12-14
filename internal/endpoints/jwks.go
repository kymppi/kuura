package endpoints

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/kymppi/kuura/internal/jwks"
)

func V1JwksHandler(logger *slog.Logger, jwkManager *jwks.JWKManager) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "max-age=600")

			ctx := r.Context()

			serviceId, err := uuid.Parse(r.PathValue("serviceId"))
			if err != nil {
				http.Error(w, "Invalid serviceId format", http.StatusBadRequest)
				return
			}

			keys, err := jwkManager.GetJWKS(ctx, serviceId)
			if err != nil {
				logger.Error("failed to get JWKS", slog.String("serviceId", serviceId.String()), slog.String("error", err.Error()))
				http.Error(w, "Failed to retrieve JWKS", http.StatusInternalServerError)
				return
			}

			if err := encode(w, r, http.StatusOK, keys); err != nil {
				logger.Error("failed to encode JWKS response", slog.String("error", err.Error()))
				http.Error(w, "Failed to encode JWKS response", http.StatusInternalServerError)
				return
			}
		},
	)
}
