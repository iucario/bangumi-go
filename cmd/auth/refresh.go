package auth

import (
	"fmt"
	"log/slog"

	"github.com/iucario/bangumi-go/api"
	"github.com/spf13/cobra"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh token",
	Run: func(cmd *cobra.Command, args []string) {
		err := api.RefreshToken(ConfigDir)
		if err != nil {
			slog.Error(fmt.Sprintln("Failed to refresh token", err))
		} else {
			slog.Info("Refresh token success")
		}
	},
}

func init() {
	authCmd.AddCommand(refreshCmd)
}
