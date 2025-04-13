package tui

import (
	"fmt"
	"strings"

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

func createCollectionText(collection api.UserSubjectCollection) string {
	text := fmt.Sprintf("[yellow]%s[-]\n%s\n\n%s\n", collection.Subject.NameCn, collection.Subject.Name, collection.Subject.ShortSummary)
	tags := strings.Join(collection.Tags, ", ")
	text += fmt.Sprintf("\nYour Tags: [green]%s[-]\n", tags)
	if collection.Rate == 0 {
		text += "Your Rate: [blue]N/A[-]\n"
	} else {
		text += fmt.Sprintf("Your Rate: [blue]%d[-]\n", collection.Rate)
	}
	text += fmt.Sprintf("Episodes Watched: %d/%d\n", collection.EpStatus, collection.Subject.Eps)
	text += fmt.Sprintf("On Aired: %s\n", collection.Subject.Date)
	text += fmt.Sprintf("User Score: %.1f\n", collection.Subject.Score)
	return text
}
