package endpoints

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/kymppi/kuura/internal/errcode"
	"github.com/kymppi/kuura/internal/errs"
	"github.com/kymppi/kuura/internal/jwks"
)

func V1JwksHandler(logger *slog.Logger, jwkManager *jwks.JWKManager) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "max-age=600")

			ctx := r.Context()

			serviceId, err := uuid.Parse(r.PathValue("serviceId"))
			if err != nil {
				handleErr(w, r, logger, errs.New(errcode.InvalidServiceId, err))
				return
			}

			keys, err := jwkManager.GetJWKS(ctx, serviceId)
			if err != nil {
				handleErr(w, r, logger, err)
				return
			}

			safeEncode(w, r, logger, http.StatusOK, keys)
		},
	)
}
