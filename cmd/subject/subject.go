package subject

import (
	"fmt"

	"github.com/iucario/bangumi-go/cmd"
	"github.com/spf13/cobra"
)

var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "Subject actions",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`Available commands:
bgm sub info <subject_id>`)
	},
}

func init() {
	cmd.RootCmd.AddCommand(subCmd)
}
