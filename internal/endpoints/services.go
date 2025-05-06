package endpoints

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/kymppi/kuura/internal/errcode"
	"github.com/kymppi/kuura/internal/errs"
	"github.com/kymppi/kuura/internal/models"
	"github.com/kymppi/kuura/internal/services"
	"github.com/kymppi/kuura/internal/users"
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

type v1ServiceUserTokens struct {
	Code         string `json:"code,omitempty"`
	SessionId    string `json:"session_id,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

func (r *v1ServiceUserTokens) Valid(ctx context.Context) (problems map[string]string) {
	problems = make(map[string]string)

	usingCode := r.Code != ""
	usingRefresh := r.SessionId != "" || r.RefreshToken != ""

	switch {
	case usingCode && usingRefresh:
		problems["code"] = "cannot provide both 'code' and 'session_id'/'refresh_token'"
	case !usingCode && !usingRefresh:
		problems["code"] = "either 'code' or 'session_id' and 'refresh_token' must be provided"
	case usingRefresh:
		if r.SessionId == "" {
			problems["session_id"] = "'session_id' cannot be empty when using refresh_token flow"
		}
		if r.RefreshToken == "" {
			problems["refresh_token"] = "'refresh_token' cannot be empty when using refresh_token flow"
		}
	}

	return problems
}

func V1_Service_UserTokens(logger *slog.Logger, userService *users.UserService) http.HandlerFunc {
	type response struct {
		AccessToken         string `json:"access_token"`
		RefreshToken        string `json:"refresh_token"`
		SessionId           string `json:"session_id"`
		AccessTokenDuration int    `json:"access_token_duration_seconds"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		pathId := r.PathValue("serviceId")
		serviceId, err := uuid.Parse(pathId)
		if err != nil {
			handleErr(w, r, logger, errs.New(errcode.InvalidServiceId, err))
			return
		}

		data, err := decodeValid[*v1ServiceUserTokens](r)
		if err != nil {
			handleErr(w, r, logger, err)
			return
		}

		var tokenInfo *users.TokenInfoForService

		if data.Code != "" {
			info, err := userService.ExchangeCodeForTokens(r.Context(), serviceId, data.Code)
			if err != nil {
				handleErr(w, r, logger, err)
				return
			}

			tokenInfo = info
		} else {
			info, err := userService.RefreshServiceAccessToken(r.Context(), serviceId, data.SessionId, data.RefreshToken)
			if err != nil {
				handleErr(w, r, logger, err)
				return
			}

			tokenInfo = info
		}

		safeEncode(w, r, logger, http.StatusOK, response{
			AccessToken:         tokenInfo.AccessToken,
			RefreshToken:        tokenInfo.RefreshToken,
			SessionId:           tokenInfo.SessionId,
			AccessTokenDuration: int(tokenInfo.AccessTokenDuration.Seconds()),
		})
	}

}
