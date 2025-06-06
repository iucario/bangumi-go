package cmd

import (
	"fmt"
	"os"

	"github.com/iucario/bangumi-go/util"
	"github.com/spf13/cobra"
)

var ConfigDir string

var RootCmd = &cobra.Command{
	Use:   "bgm",
	Short: "bgm is a command line tool for Bangumi.tv",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Bangumi CLI")
		fmt.Println("Use 'bgm help' for more information.")
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.play.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	ConfigDir = util.ConfigDir()
}
