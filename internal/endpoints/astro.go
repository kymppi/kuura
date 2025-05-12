package endpoints

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
)

func FrontendHandler(logger *slog.Logger, files embed.FS) http.Handler {
	filesystem, err := fs.Sub(files, "frontend/dist/client")
	if err != nil {
		logger.Error("Failed to access embedded frontend files", "error", err)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		})
	}

	fileServer := http.FileServer(http.FS(filesystem))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		f, err := filesystem.Open(strings.TrimPrefix(path, "/"))
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		if filepath.Ext(path) != "" {
			http.NotFound(w, r)
			return
		}

		r.URL.Path = "/index.html"
		fileServer.ServeHTTP(w, r)
	})
}
