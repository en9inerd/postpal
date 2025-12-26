package server

import (
	"log/slog"
	"net/http"

	"github.com/en9inerd/go-pkgs/httpjson"
	"github.com/yourusername/yourproject/internal/config"
)

// Example handler - replace with your own handlers
func exampleHandler(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("example handler called")
		httpjson.WriteJSON(w, httpjson.JSON{
			"message": "Hello, World!",
			"port":    cfg.Port,
		})
	}
}

// Add your handlers here
