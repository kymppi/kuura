package endpoints

import (
	"embed"
	"fmt"
	"io"
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

		if _, err := filesystem.Open(strings.TrimPrefix(path, "/")); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		if filepath.Ext(path) != "" {
			http.NotFound(w, r)
			return
		}

		indexFile, err := filesystem.Open("index.html")
		if err != nil {
			logger.Error("Failed to open index.html", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer indexFile.Close()

		stat, err := indexFile.Stat()
		if err != nil {
			logger.Error("Failed to stat index.html", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

		io.Copy(w, indexFile)
	})
}
