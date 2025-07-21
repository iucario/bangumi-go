package auth

import (
	"fmt"

	"github.com/iucario/bangumi-go/api"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show auth status",
	Run: func(cmd *cobra.Command, args []string) {
		Client := api.NewAuthClientWithConfig()
		statusFlag := Client.GetStatus()
		if statusFlag {
			fmt.Println("Auth status: OK")
		} else {
			fmt.Println("Auth status: Failed")
		}
	},
}

func init() {
	authCmd.AddCommand(statusCmd)
}
