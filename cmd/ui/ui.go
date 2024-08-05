package Ui

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
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
		userCollections, _ := list.ListUserCollection(credential.AccessToken, userInfo.Username, "all", "watch", 20, 0)
		logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
		log.Println("Starting UI command")
		TuiMain(userInfo, userCollections)
	},
}

func init() {
	cmd.RootCmd.AddCommand(uiCmd)
}

func TuiMain(userInfo auth.UserInfo, userCollections list.UserCollections) {
	app := tview.NewApplication()

	watchList := createWatchList(userCollections)
	collectionView := createCollectionView(userCollections)

	// Update subject info when an item is selected
	watchList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(userCollections.Data) {
			collection := userCollections.Data[index]
			log.Printf("Selected %s", collection.Subject.Name)
			collectionView.SetText(createCollectionText(collection))
		}
	})

	flex := tview.NewFlex().
		AddItem(watchList, 0, 1, true).
		AddItem(collectionView, 0, 2, false)
	if err := app.SetRoot(flex, true).SetFocus(flex).Run(); err != nil {
		panic(err)
	}
}

func createWatchList(userCollections list.UserCollections) *tview.List {
	watchList := tview.NewList()
	watchList.SetBorder(true).SetTitle("Watch List").SetTitleAlign(tview.AlignLeft)
	for _, collection := range userCollections.Data {
		watchList.AddItem(collection.Subject.NameCn, "", 0, nil)
	}

	// Set up keybindings for navigation
	watchList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				watchList.SetCurrentItem(watchList.GetCurrentItem() + 1)
			case 'k':
				watchList.SetCurrentItem(watchList.GetCurrentItem() - 1)
			}
		}
		return event
	})

	return watchList
}

func createCollectionView(userCollections list.UserCollections) *tview.TextView {
	subjectView := tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	subjectView.SetBorder(true).SetTitle("Subject Info").SetTitleAlign(tview.AlignLeft)
	firstCollection := userCollections.Data[0]
	subjectView.SetText(createCollectionText(firstCollection))
	return subjectView
}

func createCollectionText(collection list.UserSubjectCollection) string {
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
