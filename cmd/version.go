package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "dev" // Default version if not set during build

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information of bgm-cli",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("bgm-cli version: %s\n", Version)
	},
}
