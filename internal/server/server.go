package server

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/en9inerd/go-pkgs/httperrors"
	"github.com/en9inerd/go-pkgs/middleware"
	"github.com/en9inerd/go-pkgs/router"
	"github.com/en9inerd/postpal/internal/auth"
	"github.com/en9inerd/postpal/internal/config"
	"github.com/en9inerd/postpal/ui"
)

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' 'unsafe-hashes'")
		next.ServeHTTP(w, r)
	})
}

func NewServer(logger *slog.Logger, cfg *config.Config) (http.Handler, error) {
	authService, err := auth.NewService(
		cfg.AuthPasswordHash,
		cfg.AuthSessionSecret,
		cfg.AuthSessionMaxAge,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth service: %w", err)
	}

	templates, err := newTemplateCache()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize templates: %w", err)
	}

	r := router.New(http.NewServeMux())

	r.Use(
		SecurityHeaders,
		middleware.RealIP,
		middleware.SizeLimit(10*1024*1024),
		middleware.Recoverer(logger, false),
		middleware.GlobalThrottle(1000),
		middleware.Timeout(60*time.Second),
		middleware.Health,
	)

	staticFS, err := fs.Sub(ui.Files, "static")
	if err == nil {
		r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	}

	r.Group().Route(func(publicGroup *router.Group) {
		publicGroup.Use(Logger(logger), middleware.StripSlashes)
		registerPublicRoutes(publicGroup, logger, cfg, templates, authService)
	})

	r.Mount("/api").Route(func(apiGroup *router.Group) {
		apiGroup.Use(Logger(logger), RequireAuth(authService, logger))
		registerAPIRoutes(apiGroup, logger, cfg)
	})

	r.Group().Route(func(webGroup *router.Group) {
		webGroup.Use(Logger(logger), middleware.StripSlashes, RequireAuth(authService, logger))
		registerWebRoutes(webGroup, logger, cfg, templates)
	})

	r.NotFoundHandler(notFoundHandler(logger))

	return r, nil
}

func notFoundHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Warn("not found", "path", r.URL.Path)
		httpErr := httperrors.NewError(
			http.StatusNotFound,
			"Resource not found",
		)
		httpErr.WriteJSON(w)
	}
}
