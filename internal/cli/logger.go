package cli

import (
	"log/slog"
	"os"
)

func setupLogger(logLevel string, verbose bool) {
	level := slog.LevelInfo
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelError
	}

	if verbose {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true, // shows file:line, useful for debugging
	}

	// Create the text handler writing to Stderr
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)

	// Set as the global default.
	slog.SetDefault(logger)
}
