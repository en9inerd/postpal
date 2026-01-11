package server

import (
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/en9inerd/postpal/internal/auth"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

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

func RequireAuth(authService *auth.Service, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err != nil || cookie == nil {
				redirectToLogin(w, r, logger)
				return
			}

			valid, err := authService.ValidateSessionToken(cookie.Value)
			if err != nil || !valid {
				redirectToLogin(w, r, logger)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func redirectToLogin(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	returnURL := sanitizeReturnURL(r.URL.Path)
	if r.URL.RawQuery != "" {
		returnURL += "?" + r.URL.RawQuery
	}
	redirectPath := "/login?return=" + url.QueryEscape(returnURL)

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", redirectPath)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	http.Redirect(w, r, redirectPath, http.StatusFound)
}

func sanitizeReturnURL(path string) string {
	if path == "" || path[0] != '/' {
		return "/"
	}
	if len(path) > 1 && (path[1] == '/' || strings.Contains(path, "..")) {
		return "/"
	}
	return path
}
