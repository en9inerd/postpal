package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/en9inerd/postpal/internal/auth"
	"github.com/en9inerd/postpal/internal/config"
)

func loginHandler(logger *slog.Logger, authService *auth.Service, templates *templateCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			logger.Warn("failed to parse form", "error", err)
			renderError(w, templates, "Invalid form data")
			return
		}

		password := r.FormValue("password")
		if password == "" {
			renderError(w, templates, "Password is required")
			return
		}

		if err := authService.VerifyPassword(password); err != nil {
			logger.Warn("login failed", "ip", r.RemoteAddr)
			renderError(w, templates, "Invalid password")
			return
		}

		token, err := authService.GenerateSessionToken()
		if err != nil {
			logger.Error("failed to generate session token", "error", err)
			renderError(w, templates, "Internal server error")
			return
		}

		setSessionCookie(w, r, token, int(authService.GetSessionMaxAge().Seconds()))

		logger.Info("user logged in", "ip", r.RemoteAddr)

		returnURL := getReturnURL(r)
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", returnURL)
			return
		}

		http.Redirect(w, r, returnURL, http.StatusFound)
	}
}

func logoutHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clearSessionCookie(w, r)

		logger.Info("user logged out", "ip", r.RemoteAddr)

		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/login")
			return
		}

		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func loginPageHandler(logger *slog.Logger, cfg *config.Config, templates *templateCache, authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("session_token"); err == nil {
			if valid, err := authService.ValidateSessionToken(cookie.Value); err == nil && valid {
				returnURL := getReturnURL(r)
				http.Redirect(w, r, returnURL, http.StatusFound)
				return
			}
		}

		returnURL := getReturnURL(r)
		renderPage(w, logger, templates, "login", &templateData{
			Form:        map[string]string{"return": returnURL},
			PageTitle:   "Login - PostPal",
			PageDesc:    "Login to PostPal",
			CurrentYear: time.Now().Year(),
			Config:      cfg,
		})
	}
}

func getReturnURL(r *http.Request) string {
	returnURL := sanitizeReturnURL(r.URL.Query().Get("return"))
	if returnURL == "" {
		return "/"
	}
	return returnURL
}

func isSecure(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, token string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   isSecure(r),
		SameSite: http.SameSiteStrictMode,
	})
}

func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecure(r),
		SameSite: http.SameSiteStrictMode,
	})
}
