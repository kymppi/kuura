package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/kymppi/kuura/internal/users"
)

// func SRPVars(logger *slog.Logger, srpOptions *srp.SRPOptions) http.Handler {
// 	type SRPVarsData struct {
// 		SRPPrime     string `json:"prime"`
// 		SRPGenerator string `json:"generator"`
// 	}

// 	return http.HandlerFunc(
// 		func(w http.ResponseWriter, r *http.Request) {
// 			data := SRPVarsData{
// 				SRPPrime:     srpOptions.PrimeHex,
// 				SRPGenerator: srpOptions.Generator,
// 			}

// 			if err := encode(w, r, http.StatusOK, data); err != nil {
// 				logger.Error("failed to encode SRP vars response", slog.String("error", err.Error()))
// 				http.Error(w, "Failed to encode SRP vars response", http.StatusInternalServerError)
// 				return
// 			}
// 		},
// 	)
// }

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
	Data     string `json:"data"`
	Identity string `json:"identity"`
}

func (r *srpVerifyRequest) Valid(ctx context.Context) (problems map[string]string) {
	problems = make(map[string]string)

	if r.Data == "" {
		problems["data"] = "'data' cannot be empty"
	}
	if r.Identity == "" {
		problems["identity"] = "'identity' cannot be empty"
	}

	return problems
}

func V1_SRP_ClientVerify(logger *slog.Logger, userService *users.UserService) http.Handler {
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

			serverProof, err := userService.ClientVerify(ctx, payload.Identity, payload.Data)
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
