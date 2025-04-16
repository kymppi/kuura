package endpoints

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/kymppi/kuura/internal/errcode"
	"github.com/kymppi/kuura/internal/errs"
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

			payload, err := decodeValid[*srpClientBegin](r)
			if err != nil {
				handleErr(w, r, logger, err)
				return
			}

			value, err := userService.ClientBegin(ctx, payload.Data)
			if err != nil {
				handleErr(w, r, logger, err)
				return
			}

			data := response{
				Data: value,
			}

			safeEncode(w, r, logger, http.StatusOK, data)
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

			payload, err := decodeValid[*srpVerifyRequest](r)
			if err != nil {
				handleErr(w, r, logger, err)
				return
			}

			serverProof, uid, err := userService.ClientVerify(ctx, payload.IdentityHash, payload.Data)
			if err != nil {
				handleErr(w, r, logger, err)
				return
			}

			sessionId, refreshToken, err := userService.CreateSession(ctx, uid, uuid.MustParse(payload.TargetService))
			if err != nil {
				handleErr(w, r, logger, err)
				return
			}

			accessToken, refreshToken, serviceDomain, err := userService.CreateAccessToken(ctx, sessionId, refreshToken)
			if err != nil {
				handleErr(w, r, logger, err)
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

			safeEncode(w, r, logger, http.StatusOK, data)
		},
	)
}

func V1_User_RefreshAccessToken(logger *slog.Logger, userService *users.UserService, publicKuuraDomain string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// read session_id, refresh_token cookies
		sessionCookie, err := r.Cookie("session_id")
		if err != nil {
			handleErr(w, r, logger, errs.New(errcode.MissingCookie, errors.New("'session_id' cookie not found")).WithMetadata("cookie", "session_id"))
			return
		}
		sessionId := sessionCookie.Value

		refreshCookie, err := r.Cookie("refresh_token")
		if err != nil {
			handleErr(w, r, logger, errs.New(errcode.MissingCookie, errors.New("'refresh_token' cookie not found")).WithMetadata("cookie", "refresh_token"))
			return
		}
		refreshToken := refreshCookie.Value

		accessToken, refreshToken, serviceDomain, err := userService.CreateAccessToken(r.Context(), sessionId, refreshToken)
		if err != nil {
			handleErr(w, r, logger, err)
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

		safeEncode(w, r, logger, http.StatusOK, map[string]any{
			"success": true,
		})
	}
}
