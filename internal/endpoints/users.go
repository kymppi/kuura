package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/kymppi/kuura/internal/users"
)

type srpClientBegin struct {
	Data string `json:"data"`
}

func (r *srpClientBegin) Valid(ctx context.Context) (problems map[string]string) {
	problems = make(map[string]string)

	if r.Data == "" {
		problems["data"] = "'data' cannot be empty"
	}

	return problems
}

func V1_SRP_ClientBegin(logger *slog.Logger, userService *users.UserService) http.Handler {
	type response struct {
		Data string `json:"data"`
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			payload, problems, err := decodeValid[*srpClientBegin](r)
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

			value, err := userService.ClientBegin(ctx, payload.Data)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				logger.Error("User got an invalid challenge", slog.String("error", err.Error()))
				if encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "invalid challenge",
				}); encodeErr != nil {
					logger.Error("failed to encode validation error", slog.String("error", encodeErr.Error()))
				}
				return
			}

			data := response{
				Data: value,
			}

			if err := encode(w, r, http.StatusOK, data); err != nil {
				logger.Error("failed to encode response", slog.String("error", err.Error()))
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}
		},
	)
}

type srpVerifyRequest struct {
	Data          string `json:"data"`
	IdentityHash  string `json:"identity"`
	TargetService string `json:"target_service"`
}

func (r *srpVerifyRequest) Valid(ctx context.Context) (problems map[string]string) {
	problems = make(map[string]string)

	if r.Data == "" {
		problems["data"] = "'data' cannot be empty"
	}
	if r.IdentityHash == "" {
		problems["identity"] = "'identity' cannot be empty"
	}
	if r.TargetService == "" {
		problems["target_service"] = "'target_service' cannot be empty"
	} else {
		if _, err := uuid.Parse(r.TargetService); err != nil {
			problems["target_service"] = "'target_service' must be a valid UUID"
		}
	}

	return problems
}

func V1_SRP_ClientVerify(logger *slog.Logger, userService *users.UserService, publicKuuraDomain string) http.Handler {
	type response struct {
		Success bool   `json:"success"`
		Data    string `json:"data"`
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			payload, problems, err := decodeValid[*srpVerifyRequest](r)
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

			serverProof, uid, err := userService.ClientVerify(ctx, payload.IdentityHash, payload.Data)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				logger.Error("User got an invalid premaster", slog.String("error", err.Error()))
				if encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "invalid challenge",
				}); encodeErr != nil {
					logger.Error("failed to encode validation error", slog.String("error", encodeErr.Error()))
				}
				return
			}

			sessionId, refreshToken, err := userService.CreateSession(ctx, uid, uuid.MustParse(payload.TargetService))
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				logger.Error("Failed to create a user session", slog.String("uid", uid), slog.String("error", err.Error()))
				if encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "internal server error",
				}); encodeErr != nil {
					logger.Error("failed to encode client error", slog.String("error", encodeErr.Error()))
				}
				return
			}

			accessToken, refreshToken, serviceDomain, err := userService.CreateAccessToken(ctx, sessionId, refreshToken)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				logger.Error("Failed to create a user access token", slog.String("uid", uid), slog.String("error", err.Error()))
				if encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "internal server error",
				}); encodeErr != nil {
					logger.Error("failed to encode client error", slog.String("error", encodeErr.Error()))
				}
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "refresh_token",
				Value:    refreshToken,
				Path:     "/v1/user/access",
				MaxAge:   60 * 60 * 24 * 7, // week in seconds
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode, // path must match
				Domain:   publicKuuraDomain,
			})

			http.SetCookie(w, &http.Cookie{
				Name:     "access_token",
				Value:    accessToken,
				Path:     "/",
				MaxAge:   60 * 60, // hour in seconds
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Domain:   serviceDomain,
			})

			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    sessionId,
				Path:     "/",
				MaxAge:   60 * 60 * 24 * 7, // hour in seconds
				HttpOnly: false,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				Domain:   publicKuuraDomain,
			})

			data := response{
				Success: true,
				Data:    serverProof,
			}

			if err := encode(w, r, http.StatusOK, data); err != nil {
				logger.Error("failed to encode response", slog.String("error", err.Error()))
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}
		},
	)
}

func V1_User_RefreshAccessToken(logger *slog.Logger, userService *users.UserService, publicKuuraDomain string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// read session_id, refresh_token cookies
		sessionCookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, "'session_id' cookie not found", http.StatusBadRequest)
			return
		}
		sessionId := sessionCookie.Value

		refreshCookie, err := r.Cookie("refresh_token")
		if err != nil {
			http.Error(w, "'refresh_token' cookie not found", http.StatusBadRequest)
			return
		}
		refreshToken := refreshCookie.Value

		accessToken, refreshToken, serviceDomain, err := userService.CreateAccessToken(r.Context(), sessionId, refreshToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to refresh access token: %v", err), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Path:     "/v1/user/access",
			MaxAge:   60 * 60 * 24 * 7, // week in seconds
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode, // path must match
			Domain:   publicKuuraDomain,
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			Path:     "/",
			MaxAge:   60 * 60, // hour in seconds
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Domain:   serviceDomain,
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionId,
			Path:     "/",
			MaxAge:   60 * 60 * 24 * 7, // hour in seconds
			HttpOnly: false,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Domain:   publicKuuraDomain,
		})

		resp := map[string]any{
			"success": true,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("failed to encode response", slog.String("error", err.Error()))
		}
	}
}
