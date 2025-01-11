package logger

import (
	"log/slog"
	"os"
)

// Creates a new logger that writes to a file
func NewLogger(logPath string, logLevel string) *slog.Logger {
	options := slog.HandlerOptions{
		Level: parseLogLevel(logLevel),
	}
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		panic(err)
	}
	logger := slog.New(slog.NewJSONHandler(file, &options))
	return logger
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
