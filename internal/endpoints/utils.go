package endpoints

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/kymppi/kuura/internal/errcode"
	"github.com/kymppi/kuura/internal/errs"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"go.opentelemetry.io/otel/trace"
)

type STDErrorResponse struct {
	Message  string                     `json:"message"`
	Code     string                     `json:"code"`
	TraceID  string                     `json:"trace_id"`
	Metadata map[string]json.RawMessage `json:"metadata"`
}

func safeEncode[T any](w http.ResponseWriter, r *http.Request, logger *slog.Logger, status int, v T) {
	if err := encode(w, r, status, v); err != nil {
		logger.Error("failed to encode response", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func handleErr(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	span := trace.SpanFromContext(r.Context())
	spanContext := span.SpanContext()

	var traceId string
	if spanContext.IsValid() {
		traceId = spanContext.TraceID().String()
	}

	// in-theory the traceId should match with the customErr.TraceID but not 100% so returning both

	var customErr *errs.Error
	if errors.As(err, &customErr) {
		logger.Error("An error occurred",
			slog.String("error", customErr.Error()),
			slog.String("code", string(customErr.Code)),
		)

		errorDetail := errcode.GetErrorDetail(customErr.Code)

		safeEncode(w, r, logger, errorDetail.StatusCode, &STDErrorResponse{
			Message:  errorDetail.Description,
			Code:     string(customErr.Code),
			TraceID:  traceId,
			Metadata: customErr.Metadata,
		})
		return
	}

	safeEncode(w, r, logger, 500, &STDErrorResponse{
		Message:  "Internal server error",
		Code:     string(errcode.InternalServerError),
		TraceID:  traceId,
		Metadata: nil,
	})
}

// https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/

func encode[T any](w http.ResponseWriter, _ *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

// func decodeValid[T Validator](r *http.Request) (T, map[string]string, error) {
// 	var v T
// 	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
// 		return v, nil, fmt.Errorf("decode json: %w", err)
// 	}
// 	if problems := v.Valid(r.Context()); len(problems) > 0 {
// 		return v, problems, fmt.Errorf("invalid %T: %d problems", v, len(problems))
// 	}
// 	return v, nil, nil
// }

func decodeValid[T Validator](r *http.Request) (T, error) {
	var v T

	if r.Body == nil || r.ContentLength == 0 {
		return v, errs.New(errcode.InvalidArgumentError, fmt.Errorf("empty request body"))
	}

	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, errs.New(errcode.InternalServerError, fmt.Errorf("decode json: %w", err))
	}

	if problems := v.Valid(r.Context()); len(problems) > 0 {
		err := errs.New(errcode.InvalidArgumentError, fmt.Errorf("invalid %T: %d problems", v, len(problems)))

		problemsJSON, marshallErr := json.Marshal(problems)
		if marshallErr != nil {
			return v, errs.New(errcode.InternalServerError, fmt.Errorf("failed to marshal problems: %w", err))
		}

		err = err.WithMetadata("problems", string(problemsJSON))

		return v, err
	}

	return v, nil
}

// Validator is an object that can be validated.
type Validator interface {
	// Valid checks the object and returns any
	// problems. If len(problems) == 0 then
	// the object is valid.
	Valid(ctx context.Context) (problems map[string]string)
}

type AuthConfig struct {
	JWTIssuer string
	JWKSet    jwk.Set
}

type Client struct {
	Id             string
	Roles          []string
	ClientType     string // machine | user
	TokenExpiresAt time.Time
}

func parseToken(tokenString string, config *AuthConfig) (*Client, error) {
	token, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithKeySet(config.JWKSet),
		jwt.WithValidate(true),
		jwt.WithIssuer(config.JWTIssuer),
		jwt.WithRequiredClaim("exp"),
		jwt.WithRequiredClaim("sub"),
		jwt.WithRequiredClaim("roles"),
		jwt.WithRequiredClaim("client_type"),
	)
	if err != nil {
		return nil, fmt.Errorf("jwt validation failed: %w", err)
	}

	// exist boolean can be ignored because we already checked for it in the jwt.Parse call
	rolesInterface, _ := token.Get("roles")
	rolesSlice, ok := rolesInterface.([]any)
	if !ok {
		return nil, fmt.Errorf("invalid roles format")
	}

	roles := make([]string, len(rolesSlice))
	for i, v := range rolesSlice {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("invalid role type")
		}
		roles[i] = str
	}

	clientType, _ := token.Get("client_type")

	return &Client{
		Id:             token.Subject(),
		Roles:          roles,
		ClientType:     clientType.(string),
		TokenExpiresAt: token.Expiration(),
	}, nil
}
