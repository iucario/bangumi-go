package auth

import (
	"log"
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
		log.Fatalf("Failed to delete config directory: %s", err)
	}
}

func init() {
	authCmd.AddCommand(logoutCmd)
}
