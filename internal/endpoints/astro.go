package endpoints

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

func AstroHandler(logger *slog.Logger, files embed.FS) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			filesystem := http.FS(files)

			path := r.URL.Path

			if strings.HasSuffix(path, "/") {
				path = "index"
			}

			path = strings.TrimPrefix(path, "/")

			path = "frontend/dist/" + path

			_, err := filesystem.Open(path)
			if errors.Is(err, os.ErrNotExist) {
				path = fmt.Sprintf("%s.html", path)

				_, err = filesystem.Open(path)
				if err != nil {
					http.NotFound(w, r)
					return
				}
			}

			http.FileServer(filesystem).ServeHTTP(w, r)
		},
	)
}
