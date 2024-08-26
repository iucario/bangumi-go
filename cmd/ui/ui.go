package Ui

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd"
	"github.com/iucario/bangumi-go/cmd/auth"
	"github.com/iucario/bangumi-go/cmd/list"
	"github.com/iucario/bangumi-go/cmd/subject"
	"github.com/iucario/bangumi-go/util"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Run terminal UI",
	Run: func(cmd *cobra.Command, args []string) {
		credential, _ := auth.LoadCredential()
		userInfo, err := auth.GetUserInfo(credential.AccessToken)
		auth.AbortOnError(err)
		userCollections, _ := list.ListUserCollection(credential.AccessToken, userInfo.Username, "all", "watch", 20, 0)
		logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to open log file: %v", err))
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

func TuiMain(userInfo auth.UserInfo, userCollections api.UserCollections) {
	app := tview.NewApplication()

	pages := tview.NewPages()

	pages.AddAndSwitchToPage("help", createHelpPage(), true)

	pages.AddPage("home", createHomePage(app, userCollections), true, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case '?':
				pages.SwitchToPage("help")
			case '1':
				pages.SwitchToPage("home")
			}
		}
		return event
	})

	if err := app.SetRoot(pages, true).SetFocus(pages).Run(); err != nil {
		panic(err)
	}
}

func createHelpPage() *tview.TextView {
	text := `Welcome to Bangumi CLI UI
	Shortcuts:

	[Navigation]
	1: Go to Home
	?: Show this help
	j/up: Move up
	k/down: Move down

	[Collection]
	e: Edit collection
	v: View collection
	`
	welcomePage := tview.NewTextView().SetText(text)
	return welcomePage
}

func createHomePage(app *tview.Application, userCollections api.UserCollections) *tview.Flex {
	watchList := createWatchList(userCollections)
	collectionView := createCollectionView(userCollections)

	pages := tview.NewPages()
	pages.AddAndSwitchToPage("view", collectionView, true)

	// Update subject info when an item is selected
	watchList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(userCollections.Data) {
			collection := userCollections.Data[index]
			slog.Info(fmt.Sprintf("Selected %s", collection.Subject.Name))
			collectionView.SetText(createCollectionText(collection))
		}
	})

	homePage := tview.NewFlex().
		AddItem(watchList, 0, 1, true).
		AddItem(pages, 0, 2, false)

	homePage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'e':
				slog.Info("edit")
				index := watchList.GetCurrentItem()
				pages.AddAndSwitchToPage("edit", createEditPage(userCollections.Data[index]), true)
				_, frontPage := pages.GetFrontPage()
				app.SetFocus(frontPage)
			case 'v':
				slog.Info("view")
				pages.SwitchToPage("view")
				app.SetFocus(watchList)
			}
		}
		return event
	})

	return homePage
}

func createEditPage(collection api.UserSubjectCollection) *tview.Flex {
	form := createForm(collection)
	editPage := tview.NewFlex().
		AddItem(form, 0, 1, true)
	return editPage
}

func createWatchList(userCollections api.UserCollections) *tview.List {
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

func createCollectionView(userCollections api.UserCollections) *tview.TextView {
	subjectView := tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	subjectView.SetBorder(true).SetTitle("Subject Info").SetTitleAlign(tview.AlignLeft)
	firstCollection := userCollections.Data[0]
	subjectView.SetText(createCollectionText(firstCollection))
	return subjectView
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

func createForm(collection api.UserSubjectCollection) *tview.Form {
	// FIXME: inputs 'e' when entering edit mode. Change focus or something.
	// FIXME: should disable shortcuts when in form
	statusList := []string{"wish", "watch", "done", "onhold", "dropped"}
	status := util.IndexOfString(statusList, collection.GetStatus())
	initTags := collection.GetTags()

	form := tview.NewForm()
	form.SetBorder(true).SetTitle("Edit Collection").SetTitleAlign(tview.AlignLeft)
	form.AddInputField("Episodes watched", util.Uint32ToString(collection.EpStatus), 5, nil, func(text string) {
		epStatus, err := strconv.Atoi(text)
		if err != nil {
			slog.Error(fmt.Sprintf("invalid episode status %s", text))
		}
		collection.EpStatus = uint32(epStatus)
	})

	form.AddDropDown("Status", statusList, status, func(option string, optionIndex int) {
		slog.Debug(fmt.Sprintf("selected %s", option))
		collection.SetStatus(option)
	})
	form.AddInputField("Tags", initTags, 20, nil, func(text string) {
		// TODO: validate tags
		collection.Tags = strings.Split(text, " ")
	})
	form.AddInputField("Rate", util.Uint32ToString(collection.Rate), 2, nil, func(text string) {
		rate, err := strconv.Atoi(text)
		if err != nil {
			slog.Error(fmt.Sprintf("invalid rate %s. Must be in [0-10]", text))
		}
		rate = max(0, min(10, rate))
		collection.Rate = uint32(rate)
	})
	form.AddInputField("Comment", collection.Comment, 20, nil, func(text string) {
		collection.Comment = text
	})
	form.AddCheckbox("Private", collection.Private, func(checked bool) {
		collection.Private = checked
	})
	form.AddButton("Save", func() {
		slog.Info("save button clicked")
		slog.Info("posting collection...")
		credential, err := auth.GetCredential()
		if err != nil {
			slog.Error("login required")
			// TODO: display error messsage
		}
		subject.PostCollection(credential.AccessToken, int(collection.SubjectID), statusList[collection.Type],
			collection.Tags, collection.Comment, int(collection.Rate), collection.Private)
	})
	form.AddButton("Cancel", func() {
		slog.Info("cancel button clicked")
	})
	return form
}
