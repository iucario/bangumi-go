package main

import (
	"io"
	"log/slog"
	"os"

	"github.com/iucario/bangumi-go/cmd"
	_ "github.com/iucario/bangumi-go/cmd/auth"
	_ "github.com/iucario/bangumi-go/cmd/calendar"
	_ "github.com/iucario/bangumi-go/cmd/list"
	_ "github.com/iucario/bangumi-go/cmd/search"
	_ "github.com/iucario/bangumi-go/cmd/subject"
	_ "github.com/iucario/bangumi-go/cmd/ui"
)

func main() {
	logOutput := os.Stderr

	handler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelError,
	})
	if os.Getenv("BGM_ENV") == "dev" {
		handler = slog.NewTextHandler(logOutput, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	cmd.Execute()
}
