package cmd

import (
	"fmt"
	"os"

	"github.com/iucario/bangumi-go/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var ConfigDir string

var RootCmd = &cobra.Command{
	Use:   "bgm",
	Short: "bgm is a command line tool for Bangumi.tv",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	cmd, _, err := RootCmd.Find(os.Args[1:])
	// default cmd if no cmd is given
	if err == nil && cmd.Use == RootCmd.Use && cmd.Flags().Parse(os.Args[1:]) != pflag.ErrHelp {
		args := append([]string{"ui"}, os.Args[1:]...)
		RootCmd.SetArgs(args)
	}

	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	ConfigDir = util.ConfigDir()
}
