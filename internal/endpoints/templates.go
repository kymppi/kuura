package endpoints

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/kymppi/kuura/internal/srp"
)

func LoginPage(logger *slog.Logger, tmpl *template.Template, srpOptions *srp.SRPOptions) http.Handler {
	type LoginPageData struct {
		SRPPrime     string
		SRPGenerator string
	}

	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			data := LoginPageData{
				SRPPrime:     srpOptions.PrimeHex,
				SRPGenerator: srpOptions.Generator,
			}

			err := tmpl.ExecuteTemplate(w, "login.tmpl", data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		},
	)
}
