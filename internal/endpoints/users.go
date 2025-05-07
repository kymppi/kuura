package endpoints

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/kymppi/kuura/internal/constants"
	"github.com/kymppi/kuura/internal/errcode"
	"github.com/kymppi/kuura/internal/errs"
	"github.com/kymppi/kuura/internal/jwks"
	"github.com/kymppi/kuura/internal/users"
	"github.com/lestrrat-go/jwx/v2/jwt"
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

			sessionId, initialRefreshToken, err := userService.CreateSession(ctx, uid, uuid.MustParse(payload.TargetService))
			if err != nil {
				handleErr(w, r, logger, err)
				return
			}

			tokenInfo, err := userService.CreateAccessToken(ctx, sessionId, initialRefreshToken)
			if err != nil {
				handleErr(w, r, logger, err)
				return
			}

			setAuthCookies(w, sessionId, tokenInfo, publicKuuraDomain)

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
		sessionCookie, err := r.Cookie(constants.SESSION_COOKIE)
		if err != nil {
			handleErr(w, r, logger, errs.New(errcode.MissingCookie, fmt.Errorf("'%s' cookie not found", constants.SESSION_COOKIE)).WithMetadata("cookie", constants.SESSION_COOKIE))
			return
		}
		sessionId := sessionCookie.Value

		refreshCookie, err := r.Cookie(constants.REFRESH_TOKEN_COOKIE)
		if err != nil {
			handleErr(w, r, logger, errs.New(errcode.MissingCookie, fmt.Errorf("'%s' cookie not found", constants.REFRESH_TOKEN_COOKIE)).WithMetadata("cookie", constants.REFRESH_TOKEN_COOKIE))
			return
		}
		initialRefreshToken := refreshCookie.Value

		tokenInfo, err := userService.CreateAccessToken(r.Context(), sessionId, initialRefreshToken)
		if err != nil {
			handleErr(w, r, logger, err)
			return
		}

		setAuthCookies(w, sessionId, tokenInfo, publicKuuraDomain)

		safeEncode(w, r, logger, http.StatusOK, map[string]any{
			"success": true,
		})
	}
}

func V1_ME(logger *slog.Logger, users *users.UserService, jwkManager *jwks.JWKManager, jwtIssuer string) http.Handler {
	type response struct {
		Id          string `json:"id"`
		Username    string `json:"username"`
		LastLoginAt string `json:"last_login_at"`
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			accessCookie, err := r.Cookie(constants.KUURA_ACCESS_TOKEN_COOKIE)
			if err != nil {
				handleErr(w, r, logger, errs.New(errcode.Unauthorized, fmt.Errorf("'%s' cookie not found", constants.KUURA_ACCESS_TOKEN_COOKIE)))
				return
			}
			token := accessCookie.Value

			serviceId, err := extractServiceIdFromToken(token)
			if err != nil {
				handleErr(w, r, logger, errs.New(errcode.Unauthorized, fmt.Errorf("failed to extract serviceId from token: %w", err)))
				return
			} else if serviceId == nil {
				handleErr(w, r, logger, errs.New(errcode.Unauthorized, fmt.Errorf("failed to extract serviceId from token without error")))
				return
			}

			jwkSet, err := jwkManager.GetJWKS(ctx, *serviceId)
			if err != nil {
				handleErr(w, r, logger, errs.New(errcode.Unauthorized, fmt.Errorf("failed to get JWKS: %w", err)))
				return
			}

			client, err := parseToken(token, &AuthConfig{
				JWTIssuer: jwtIssuer,
				JWKSet:    jwkSet,
			})
			if err != nil {
				handleErr(w, r, logger, errs.New(errcode.Unauthorized, err))
				return
			}

			user, err := users.GetUser(ctx, client.Id)
			if err != nil {
				handleErr(w, r, logger, err)
				return
			}

			safeEncode(w, r, logger, http.StatusOK, response{
				Id:          user.Id,
				Username:    user.Username,
				LastLoginAt: user.LastLoginAt.UTC().Format("2006-01-02T15:04:05Z"),
			})
		},
	)
}

// extracts the serviceId from a JWT token without fully validating it
func extractServiceIdFromToken(tokenString string) (*uuid.UUID, error) {
	token, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithVerify(false),   // Skip signature verification
		jwt.WithValidate(false), // Skip validation
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	serviceId, ok := token.Get("service_id")
	if !ok {
		return nil, fmt.Errorf("serviceId claim not found in token")
	}

	serviceIdStr, ok := serviceId.(string)
	if !ok || serviceIdStr == "" {
		return nil, fmt.Errorf("serviceId claim is not a valid string")
	}

	parsedUUID, err := uuid.Parse(serviceIdStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse serviceId: %s", err)
	}

	return &parsedUUID, nil
}

func setAuthCookies(w http.ResponseWriter, sessionId string, tokenInfo *users.TokenInfo, publicKuuraDomain string) {
	// refresh token
	http.SetCookie(w, &http.Cookie{
		Name:     constants.REFRESH_TOKEN_COOKIE,
		Value:    tokenInfo.RefreshToken,
		Path:     "/v1/user/access",
		MaxAge:   60 * 60 * 24 * 7, // week in seconds
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode, // path must match
		Domain:   publicKuuraDomain,
	})

	// session
	http.SetCookie(w, &http.Cookie{
		Name:     constants.SESSION_COOKIE,
		Value:    sessionId,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 30, // month in seconds
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Domain:   publicKuuraDomain,
	})

	// service specific access token
	http.SetCookie(w, &http.Cookie{
		Name:     tokenInfo.ServiceAccessCookieName,
		Value:    tokenInfo.AccessToken,
		Path:     "/",
		MaxAge:   int(tokenInfo.AccessTokenDuration.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Domain:   tokenInfo.ServiceDomain,
	})
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
		data, err := decodeValid[*v1ServiceUserTokens](r)
		if err != nil {
			handleErr(w, r, logger, err)
			return
		}

		var tokenInfo *users.TokenInfo

		if data.Code != "" {
			info, err := userService.CreateAccessTokenUsingCode(r.Context(), data.Code)
			if err != nil {
				handleErr(w, r, logger, err)
				return
			}

			tokenInfo = info
		} else {
			info, err := userService.CreateAccessToken(r.Context(), data.SessionId, data.RefreshToken)
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
