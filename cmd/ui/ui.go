package Ui

import (
	"github.com/iucario/bangumi-go/cmd"
	"github.com/iucario/bangumi-go/cmd/auth"
	"github.com/iucario/bangumi-go/cmd/list"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Run terminal UI",
	Run: func(cmd *cobra.Command, args []string) {
		credential, _ := auth.LoadCredential()
		userInfo, err := auth.GetUserInfo(credential.AccessToken)
		auth.Check(err)
		userCollections, _ := list.ListAnimeCollection(credential.AccessToken, userInfo.Username, 3, 20, 0)
		TuiMain(userCollections)
	},
}

func init() {
	cmd.RootCmd.AddCommand(uiCmd)
}

func TuiMain(userCollections list.UserCollections) {
	app := tview.NewApplication()
	flex := tview.NewFlex().
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Left"), 0, 1, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Top 4 rows"), 4, 1, false).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Middle 2*left_width"), 0, 3, false).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("Bottom 4 rows"), 4, 1, false), 0, 2, false)
	if err := app.SetRoot(flex, true).SetFocus(flex).Run(); err != nil {
		panic(err)
	}
}
