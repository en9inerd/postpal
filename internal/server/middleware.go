package server

import (
	"log/slog"
	"net"
	"net/http"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Logger middleware logs HTTP requests
func Logger(l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()

			next.ServeHTTP(sw, r)

			duration := time.Since(start)

			remoteIP := "-"
			if r.RemoteAddr != "" {
				if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil && host != "" {
					remoteIP = host
				}
			}

			l.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"ip", remoteIP,
				"status", sw.status,
				"duration", duration,
			)
		})
	}
}
