package config

import (
	"log/slog"
	"os"
)

func Logger(isLevelDebug bool) {
	var level slog.Level
	if isLevelDebug {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
