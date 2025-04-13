package tui

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd/list"
	"github.com/iucario/bangumi-go/cmd/subject"
	"github.com/iucario/bangumi-go/util"
	"github.com/rivo/tview"
)

type App struct {
	*tview.Application
	Pages *tview.Pages
	User  *api.User
}

func NewApp(user *api.User) *App {
	return &App{
		Application: tview.NewApplication(),
		Pages:       tview.NewPages(),
		User:        user,
	}
}

func (a *App) Run() error {
	userInfo, err := a.User.GetUserInfo()
	if err != nil {
		fmt.Println("Login required. Please run `bgm auth login` first.")
		return err
	}
	options := list.UserListOptions{
		SubjectType:    "all",
		Username:       userInfo.Username,
		CollectionType: "all",
		Limit:          20,
		Offset:         0,
	}
	userCollections, err := list.ListUserCollection(a.User.Client, options)
	if err != nil {
		return err
	}

	a.Pages.AddPage("help", createHelpPage(), true, false)
	a.Pages.AddAndSwitchToPage("home", newHomePage(a.Application, *userCollections), true)

	a.Application.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case '?':
				a.Pages.SwitchToPage("help")
			case '1':
				a.Pages.SwitchToPage("home")
			}
		}
		return event
	})

	if err := a.Application.SetRoot(a.Pages, true).SetFocus(a.Pages).Run(); err != nil {
		panic(err)
	}
	return nil
}

// newHomePage creates the home page of the TUI application.
// Left side shows the watch list, right side shows the collection view.
func newHomePage(app *tview.Application, userCollections api.UserCollections) *tview.Flex {
	watchList := newWatchList(userCollections)
	collectionView := newCollectionView(userCollections)

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

// newWatchList creates a list of user collections.
func newWatchList(userCollections api.UserCollections) *tview.List {
	watchList := tview.NewList()
	watchList.SetBorder(true).SetTitle("Watch List").SetTitleAlign(tview.AlignLeft)
	for _, collection := range userCollections.Data {
		name := collection.Subject.NameCn
		if name == "" {
			name = collection.Subject.Name
		}
		watchList.AddItem(name, "", 0, nil)
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

// newCollectionView creates a view for the selected collection.
func newCollectionView(userCollections api.UserCollections) *tview.TextView {
	subjectView := tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	subjectView.SetBorder(true).SetTitle("Subject Info").SetTitleAlign(tview.AlignLeft)
	firstCollection := userCollections.Data[0]
	subjectView.SetText(createCollectionText(firstCollection))
	return subjectView
}

func createEditPage(collection api.UserSubjectCollection) *tview.Flex {
	form := createForm(collection)
	editPage := tview.NewFlex().
		AddItem(form, 0, 1, true)
	return editPage
}

func createForm(collection api.UserSubjectCollection) *tview.Form {
	// FIXME: inputs 'e' when entering edit mode. Change focus or something.
	// FIXME: should disable shortcuts when in form
	statusList := []string{"wish", "done", "watch", "onhold", "dropped"}
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
		credential, err := api.GetCredential()
		if err != nil {
			slog.Error("login required")
			// TODO: display error messsage
		}
		err = subject.PostCollection(credential.AccessToken, int(collection.SubjectID), statusList[collection.Type-1],
			collection.Tags, collection.Comment, int(collection.Rate), collection.Private)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to post collection: %v", err))
		}
		subject.WatchToEpisode(credential.AccessToken, int(collection.SubjectID), int(collection.EpStatus))
	})
	form.AddButton("Cancel", func() {
		slog.Info("cancel button clicked")
		// TODO: implement cancel action

	})
	return form
}
