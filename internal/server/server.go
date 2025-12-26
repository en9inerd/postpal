package server

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/en9inerd/go-pkgs/middleware"
	"github.com/en9inerd/go-pkgs/router"
	"github.com/yourusername/yourproject/internal/config"
	"github.com/yourusername/yourproject/ui"
)

// SecurityHeaders adds security headers to responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' 'unsafe-hashes'")
		next.ServeHTTP(w, r)
	})
}

// NewServer creates and configures a new HTTP server handler
func NewServer(
	logger *slog.Logger,
	cfg *config.Config,
) (http.Handler, error) {
	r := router.New(http.NewServeMux())

	r.Use(
		SecurityHeaders,
		middleware.RealIP,
		middleware.Recoverer(logger, false),
		middleware.GlobalThrottle(1000),
		middleware.Timeout(60*time.Second),
		middleware.Health,
	)

	// Serve static files if UI is embedded
	staticFS, err := fs.Sub(ui.Files, "static")
	if err == nil {
		r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	}

	// API routes
	r.Mount("/api").Route(func(apiGroup *router.Group) {
		registerAPIRoutes(apiGroup, logger, cfg)
	})

	// Web routes (if using templates)
	r.Group().Route(func(webGroup *router.Group) {
		registerWebRoutes(webGroup, logger, cfg)
	})

	// 404 handler
	r.NotFoundHandler(notFoundHandler(logger))

	return r, nil
}

// registerAPIRoutes registers API endpoints
func registerAPIRoutes(
	apiGroup *router.Group,
	logger *slog.Logger,
	cfg *config.Config,
) {
	apiGroup.Use(Logger(logger))
	// Add your API routes here
	// Example:
	// apiGroup.HandleFunc("GET /health", healthHandler(logger))
	// apiGroup.HandleFunc("POST /users", createUserHandler(logger, cfg))
}

// registerWebRoutes registers web page routes
func registerWebRoutes(
	webGroup *router.Group,
	logger *slog.Logger,
	cfg *config.Config,
) {
	webGroup.Use(Logger(logger), middleware.StripSlashes)
	// Add your web routes here
	// Example:
	// webGroup.HandleFunc("GET /", homePage(logger, cfg))
}

// notFoundHandler handles 404 requests
func notFoundHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Warn("not found", "path", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Not Found")
	}
}
