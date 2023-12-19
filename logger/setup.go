package logger

import (
	"log/slog"
	"os"
)

func Setup() {
	if os.Getenv("LOCAL_LOGGER") == "true" {
		logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
		slog.SetDefault(logger)
	} else {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		slog.SetDefault(logger)
	}
}
