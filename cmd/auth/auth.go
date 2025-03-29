package auth

import (
	"fmt"

	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd"
	"github.com/spf13/cobra"
)

var (
	ConfigDir string
	Client    *api.AuthClient
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Auth commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`Available commands: 
bgm auth login
bgm auth logout
bgm auth status`)
	},
}

func init() {
	cmd.RootCmd.AddCommand(authCmd)
	ConfigDir = cmd.ConfigDir

	credential, _ := api.LoadCredential()
	Client = api.NewAuthClient(credential.AccessToken)
}
