package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/iucario/bangumi-go/cmd"
	_ "github.com/iucario/bangumi-go/cmd/auth"
	_ "github.com/iucario/bangumi-go/cmd/calendar"
	_ "github.com/iucario/bangumi-go/cmd/list"
	_ "github.com/iucario/bangumi-go/cmd/subject"
	_ "github.com/iucario/bangumi-go/cmd/ui"
)

func main() {
	logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to open log file: %v", err))
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			slog.Error(fmt.Sprintf("Failed to close log file: %v", err))
		}
	}()
	logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	cmd.Execute()
}
