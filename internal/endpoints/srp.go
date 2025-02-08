package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"

	"github.com/kymppi/kuura/internal/srp"
	"github.com/kymppi/kuura/internal/users"
)

func SRPVars(logger *slog.Logger, srpOptions *srp.SRPOptions) http.Handler {
	type SRPVarsData struct {
		SRPPrime     string `json:"prime"`
		SRPGenerator string `json:"generator"`
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			data := SRPVarsData{
				SRPPrime:     srpOptions.PrimeHex,
				SRPGenerator: srpOptions.Generator,
			}

			if err := encode(w, r, http.StatusOK, data); err != nil {
				logger.Error("failed to encode SRP vars response", slog.String("error", err.Error()))
				http.Error(w, "Failed to encode SRP vars response", http.StatusInternalServerError)
				return
			}
		},
	)
}

type srpChallengeRequest struct {
	I string `json:"I"`
	A string `json:"A"`
}

func (r *srpChallengeRequest) Valid(ctx context.Context) (problems map[string]string) {
	problems = make(map[string]string)

	if r.I == "" {
		problems["I"] = "'I' cannot be empty"
	}
	if r.A == "" {
		problems["A"] = "'A' cannot be empty"
	} else {
		if _, ok := new(big.Int).SetString(r.A, 16); !ok {
			problems["A"] = "'A' is invalid"
		}
	}

	return problems
}

func V1_SRPChallenge(logger *slog.Logger, userService *users.UserService) http.Handler {
	type response struct {
		Salt string `json:"s"`
		B    string `json:"B"`
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			payload, problems, err := decodeValid[*srpChallengeRequest](r)
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

			/*
				1. s, v = lookup(I)
				2. gen b and B
				3. calculate premaster and store in db with 5min TTL
				4. return s and B
			*/

			A, _ := new(big.Int).SetString(payload.A, 16) // validation already checked for errors
			salt, B, err := userService.SRPChallenge(ctx, payload.I, A)
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
				Salt: salt,
				B:    B.String(),
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
	Premaster string `json:"premaster"`
	I         string `json:"I"`
}

func (r *srpVerifyRequest) Valid(ctx context.Context) (problems map[string]string) {
	problems = make(map[string]string)

	if r.I == "" {
		problems["I"] = "'I' cannot be empty"
	}
	if r.Premaster == "" {
		problems["premaster"] = "'premaster' cannot be empty"
	}

	return problems
}

func V1_SRPVerify(logger *slog.Logger, userService *users.UserService) http.Handler {
	type response struct {
		Success bool `json:"success"`
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

			err = userService.SRPVerify(ctx, payload.I, payload.Premaster)
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
			}

			if err := encode(w, r, http.StatusOK, data); err != nil {
				logger.Error("failed to encode response", slog.String("error", err.Error()))
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}
		},
	)
}
