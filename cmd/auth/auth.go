package auth

import (
	"fmt"
	"log/slog"

	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd"
	"github.com/spf13/cobra"
)

var (
	ConfigDir string
	Client    *api.AuthClient = api.NewAuthClient("")
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Auth commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`Available commands: 
bgm auth login
bgm auth logout
bgm auth status
bgm auth refresh`)
	},
}

func init() {
	cmd.RootCmd.AddCommand(authCmd)
	ConfigDir = cmd.ConfigDir
	initializeAuthClient()
}

func initializeAuthClient() {
	credential, err := api.GetCredential()
	if err != nil || credential == nil {
		slog.Error("Failed to get credential", "error", err)
		return
	}

	Client = api.NewAuthClient(credential.AccessToken)
}
