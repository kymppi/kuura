package middleware

import (
	"log/slog"
	"net/http"
)

type statusTrackingWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusTrackingWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func LoggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stw := &statusTrackingWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(stw, r)

		logger.Info("HTTP Request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", stw.status),
			slog.String("ip", r.RemoteAddr),
		)
	})
}
