package log

import (
	"io"
	"log/slog"
	"os"
)

// NewLogger returns a logger that only emits logs when verbose is true.
func NewLogger(verbose bool) *slog.Logger {
	var out io.Writer
	var level slog.Level

	if verbose {
		out = os.Stderr
		level = slog.LevelDebug
	} else {
		out = io.Discard
		level = slog.LevelError
	}

	handler := slog.NewTextHandler(out, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler)
}
