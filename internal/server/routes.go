package server

import (
	"log/slog"

	"github.com/en9inerd/go-pkgs/router"
	"github.com/en9inerd/postpal/internal/auth"
	"github.com/en9inerd/postpal/internal/config"
)

func registerAPIRoutes(apiGroup *router.Group, logger *slog.Logger, cfg *config.Config) {
}

func registerWebRoutes(webGroup *router.Group, logger *slog.Logger, cfg *config.Config, templates *templateCache) {
}

func registerPublicRoutes(publicGroup *router.Group, logger *slog.Logger, cfg *config.Config, templates *templateCache, authService *auth.Service) {
	publicGroup.HandleFunc("GET /login", loginPageHandler(logger, cfg, templates, authService))
	publicGroup.HandleFunc("POST /login", loginHandler(logger, authService, templates))
	publicGroup.HandleFunc("POST /logout", logoutHandler(logger))
}
