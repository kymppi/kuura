package endpoints

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

func AstroHandler(logger *slog.Logger, files embed.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filesystem, err := fs.Sub(files, "frontend-v2/dist")
		if err != nil {
			http.Error(w, "Failed to access the directory", http.StatusInternalServerError)
			return
		}

		path := r.URL.Path

		if strings.HasSuffix(path, "/") {
			path = "index"
		}

		path = strings.TrimPrefix(path, "/")

		_, err = filesystem.Open(path)
		if errors.Is(err, os.ErrNotExist) {
			path = fmt.Sprintf("%s.html", path)

			_, err = filesystem.Open(path)
			if err != nil {
				http.NotFound(w, r)
				return
			}
		}

		http.FileServer(http.FS(filesystem)).ServeHTTP(w, r)
	})
}
