package endpoints

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/kymppi/kuura/internal/errcode"
	"github.com/kymppi/kuura/internal/errs"
	"github.com/kymppi/kuura/internal/services"
)

func V1_ServiceInfo(logger *slog.Logger, serviceManager *services.ServiceManager) http.Handler {
	type response struct {
		Name         string `json:"name"`
		ContactName  string `json:"contact"`
		ContactEmail string `json:"contact_email"`
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "max-age=600")

			ctx := r.Context()

			serviceId, err := uuid.Parse(r.PathValue("serviceId"))
			if err != nil {
				handleErr(w, r, logger, errs.New(errcode.InvalidServiceId, err))
				return
			}

			svc, err := serviceManager.GetService(ctx, serviceId)
			if err != nil {
				handleErr(w, r, logger, err)
				return
			}

			safeEncode(w, r, logger, http.StatusOK, response{
				Name:         svc.Name,
				ContactName:  svc.ContactName,
				ContactEmail: svc.ContactEmail,
			})
		},
	)
}
