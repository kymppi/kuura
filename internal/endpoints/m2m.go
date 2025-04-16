package endpoints

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/kymppi/kuura/internal/m2m"
)

type v1CreateM2MSessionRequest struct {
	SubjectId string `json:"subject_id"`
	Template  string `json:"template"`
	ServiceId string `json:"service_id"`
}

func (r *v1CreateM2MSessionRequest) Valid(ctx context.Context) (problems map[string]string) {
	problems = make(map[string]string)

	if r.SubjectId == "" {
		problems["subject_id"] = "'subject_id' cannot be empty"
	}
	if r.Template == "" {
		problems["template"] = "'template' cannot be empty"
	}
	if r.ServiceId == "" {
		problems["service_id"] = "'service_id' cannot be empty"
	} else {
		if _, err := uuid.Parse(r.ServiceId); err != nil {
			problems["service_id"] = "'service_id' must be a valid UUID"
		}
	}

	return problems
}

func V1CreateM2MSession(logger *slog.Logger, m2mService *m2m.M2MService) http.HandlerFunc {
	type response struct {
		SessionId    string `json:"session_id"`
		InitialToken string `json:"token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		data, err := decodeValid[*v1CreateM2MSessionRequest](r)
		if err != nil {
			handleErr(w, r, logger, err)
			return
		}

		sessionId, initialToken, err := m2mService.CreateSession(r.Context(), uuid.MustParse(data.ServiceId), data.SubjectId, data.Template)
		if err != nil {
			handleErr(w, r, logger, err)
			return
		}

		safeEncode(w, r, logger, http.StatusCreated, response{
			SessionId:    sessionId,
			InitialToken: initialToken,
		})
	}
}

type v1M2MRefreshAccessToken struct {
	SessionId    string `json:"session_id"`
	RefreshToken string `json:"refresh_token"`
}

func (r *v1M2MRefreshAccessToken) Valid(ctx context.Context) (problems map[string]string) {
	problems = make(map[string]string)

	if r.SessionId == "" {
		problems["session_id"] = "'session_id' cannot be empty"
	}
	if r.RefreshToken == "" {
		problems["refresh_token"] = "'refresh_token' cannot be empty"
	}

	return problems
}

func V1M2MRefreshAccessToken(logger *slog.Logger, m2mService *m2m.M2MService) http.HandlerFunc {
	type response struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		data, err := decodeValid[*v1M2MRefreshAccessToken](r)
		if err != nil {
			handleErr(w, r, logger, err)
			return
		}

		accessToken, refreshToken, err := m2mService.CreateAccessToken(r.Context(), data.SessionId, data.RefreshToken)
		if err != nil {
			handleErr(w, r, logger, err)
			return
		}

		safeEncode(w, r, logger, http.StatusOK, response{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		})
	}
}
