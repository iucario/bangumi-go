package auth

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh token",
	Run: func(cmd *cobra.Command, args []string) {
		credential, err := client.RefreshToken()
		if err != nil {
			fmt.Println("Failed to refresh token", err)
		} else {
			expiration := credential.ExpiresIn
			expirationTime := time.Now().Add(time.Duration(expiration) * time.Second)
			date := expirationTime.Format("2006-01-02 15:04:05")
			fmt.Printf("Refresh token success. New token expires at %s\n", date)
		}
	},
}

func init() {
	authCmd.AddCommand(refreshCmd)
}
