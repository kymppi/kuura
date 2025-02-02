package endpoints

import (
	"log/slog"
	"net/http"

	"github.com/kymppi/kuura/internal/srp"
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
