package tui

import (
	"fmt"

	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd"
	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Run terminal UI",
	Run: func(cmd *cobra.Command, args []string) {
		authClient := api.NewAuthClientWithConfig()
		user := api.NewUser(authClient)
		if user == nil {
			fmt.Println("Please login. Run `bgm auth login`")
			return
		}

		app := NewApp(user)
		err := app.Run()
		if err != nil {
			fmt.Println("Error running app:", err)
			return
		}
	},
}

func init() {
	cmd.RootCmd.AddCommand(uiCmd)
}
