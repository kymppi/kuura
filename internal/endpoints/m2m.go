package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
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
		data, problems, err := decodeValid[*v1CreateM2MSessionRequest](r)
		if err != nil {
			if problems != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				if encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
					"error":    "validation failed",
					"problems": problems,
				}); encodeErr != nil {
					logger.Error("failed to encode validation error", slog.String("error", encodeErr.Error()))
				}
			} else {
				http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
			}
			return
		}

		sessionId, initialToken, err := m2mService.CreateSession(r.Context(), uuid.MustParse(data.ServiceId), data.SubjectId, data.Template)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create session: %v", err), http.StatusInternalServerError)
			return
		}

		resp := response{
			SessionId:    sessionId,
			InitialToken: initialToken,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("failed to encode response", slog.String("error", err.Error()))
		}
	}
}
