package auth

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Auth commands",
	Run: func(cmd *cobra.Command, args []string) {
		deleteCredential()
		log.Println("Logout success.")
	},
}

func deleteCredential() {
	err := os.RemoveAll(ConfigDir)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to delete config directory: %s", err))
	}
}

func init() {
	authCmd.AddCommand(logoutCmd)
}
