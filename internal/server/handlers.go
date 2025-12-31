package server

import (
	"log/slog"
	"net/http"

	"github.com/en9inerd/go-pkgs/httpjson"
	"github.com/en9inerd/go-pkgs/httperrors"
	"github.com/en9inerd/postpal/internal/config"
)

// Example handler - replace with your own handlers
func exampleHandler(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("example handler called")
		httpjson.EncodeJSON(w, http.StatusOK, httpjson.JSON{
			"message": "Hello, World!",
			"port":    cfg.Port,
		})
	}
}

// Example POST handler showing proper request parsing and error handling
func examplePostHandler(logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Example request struct
		type ExampleRequest struct {
			Message string `json:"message"`
		}

		var req ExampleRequest
		if err := httpjson.DecodeJSON(r, &req); err != nil {
			httpErr := httperrors.NewError(
				http.StatusBadRequest,
				"Invalid request body",
			)
			httpErr.WriteJSON(w)
			return
		}

		// Process request...
		logger.Info("example POST handler called", "message", req.Message)

		httpjson.EncodeJSON(w, http.StatusOK, httpjson.JSON{
			"status":  "success",
			"message": req.Message,
		})
	}
}

// Add your handlers here
