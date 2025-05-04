package endpoints

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/kymppi/kuura/internal/errcode"
	"github.com/kymppi/kuura/internal/errs"
	"github.com/kymppi/kuura/internal/models"
	"github.com/kymppi/kuura/internal/services"
)

const KUURA_SERVICE_ID_PATH = "kuura"

func V1_ServiceInfo(logger *slog.Logger, serviceManager *services.ServiceManager) http.Handler {
	type response struct {
		Id            string `json:"id"`
		Name          string `json:"name"`
		ContactName   string `json:"contact"`
		ContactEmail  string `json:"contact_email"`
		LoginRedirect string `json:"login_redirect"`
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			var service *models.AppService
			pathId := r.PathValue("serviceId")

			if pathId == KUURA_SERVICE_ID_PATH {
				svc, err := serviceManager.GetInternalKuuraService(ctx)
				if err != nil {
					handleErr(w, r, logger, err)
					return
				}

				service = svc
			} else {
				serviceId, err := uuid.Parse(pathId)
				if err != nil {
					handleErr(w, r, logger, errs.New(errcode.InvalidServiceId, err))
					return
				}

				svc, err := serviceManager.GetService(ctx, serviceId)
				if err != nil {
					handleErr(w, r, logger, err)
					return
				}

				service = svc
			}

			safeEncode(w, r, logger, http.StatusOK, response{
				Id:            service.Id.String(),
				Name:          service.Name,
				ContactName:   service.ContactName,
				ContactEmail:  service.ContactEmail,
				LoginRedirect: service.LoginRedirect,
			})
		},
	)
}
