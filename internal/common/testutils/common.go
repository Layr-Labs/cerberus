package testutils

import (
	"log/slog"
	"os"
)

func GetTestLogger() *slog.Logger {
	handler := slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug},
	)
	return slog.New(handler)
}
