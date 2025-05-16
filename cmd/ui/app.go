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

	"github.com/iucario/bangumi-go/internal/ui"
)

type App struct {
	*tview.Application
	Pages           *tview.Pages
	User            *api.User
	UserCollections []api.UserSubjectCollection
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
	a.UserCollections = userCollections.Data

	a.Pages.AddAndSwitchToPage("home", a.NewHomePage(), true)
	a.Pages.AddPage("help", a.NewHelpPage(), true, false)

	if err := a.Application.SetRoot(a.Pages, true).SetFocus(a.Pages).Run(); err != nil {
		panic(err)
	}
	return nil
}

// newHomePage creates the home page of the TUI application.
// Left side shows the watch list, right side shows the collection view.
func (a *App) NewHomePage() *tview.Flex {
	watchList := a.NewWatchList()
	collectionView := newCollectionView(&api.UserCollections{Data: a.UserCollections})

	pages := tview.NewPages()
	pages.AddAndSwitchToPage("view", collectionView, true)

	// Update subject info when an item is selected
	watchList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(a.UserCollections) {
			collection := a.UserCollections[index]
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
				modal := a.NewEditModel(a.UserCollections[index])
				a.Pages.AddPage("edit", modal, true, true) // Add modal as an overlay
				a.SetFocus(modal)                          // Set focus to the modal
			case 'v':
				slog.Info("view")
				pages.SwitchToPage("view")
				a.SetFocus(watchList)
			}
		}
		return event
	})

	return homePage
}

func (a *App) NewHelpPage() *tview.TextView {
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

	welcomePage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case '1':
				a.Pages.SwitchToPage("home")
			}
		}
		return event
	})
	return welcomePage
}

// newWatchList creates a list of user collections titles.
func (a *App) NewWatchList() *tview.List {
	watchList := tview.NewList()
	watchList.SetBorder(true).SetTitle("Watch List").SetTitleAlign(tview.AlignLeft)
	for _, collection := range a.UserCollections {
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
			case '?':
				a.Pages.SwitchToPage("help")
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
func newCollectionView(userCollections *api.UserCollections) *tview.TextView {
	subjectView := tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	subjectView.SetBorder(true).SetTitle("Subject Info").SetTitleAlign(tview.AlignLeft)
	firstCollection := userCollections.Data[0]
	subjectView.SetText(createCollectionText(firstCollection))
	return subjectView
}

func (a *App) NewEditModel(collection api.UserSubjectCollection) *ui.Modal {
	closeFn := func() {
		a.Pages.RemovePage("edit")
		a.SetFocus(a.Pages) // Restore focus to the main page
	}
	form := createForm(collection, a, closeFn)
	modal := ui.NewModalForm("Edit Collection", form)

	// Set input capture at the form level to catch Esc
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			closeFn()
			return nil // Prevent event from propagating
		}
		return event
	})

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex == -1 {
			slog.Info("modal closed")
			closeFn()
		} else {
			slog.Info(fmt.Sprintf("button %d clicked", buttonIndex))
			slog.Info(fmt.Sprintf("button %s clicked", buttonLabel))
		}
	})

	return modal
}

func createForm(collection api.UserSubjectCollection, a *App, closeFn func()) *tview.Form {
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

		// Reorder the list and update the data in the watch list
		updatedIndex := indexOfCollection(a.UserCollections, collection.SubjectID)
		if updatedIndex != -1 {
			newCollections := reorderedSlice(a.UserCollections, updatedIndex)
			a.UserCollections = newCollections
			a.UserCollections[0] = collection
			watchList := a.NewWatchList()
			watchList.SetCurrentItem(0)
			a.Pages.RemovePage("home")
			a.Pages.AddAndSwitchToPage("home", a.NewHomePage(), true)
		}

		// Back to home page
		closeFn()
	})
	form.AddButton("Cancel", func() {
		slog.Info("cancel button clicked")
		closeFn()
	})
	return form
}

// indexOfCollection finds the index of a collection in the user collections by SubjectID.
func indexOfCollection(collections []api.UserSubjectCollection, subjectID uint32) int {
	for i, collection := range collections {
		if collection.SubjectID == subjectID {
			return i
		}
	}
	return -1 // Return -1 if not found
}

// Move the item at index to the front of the slice
func reorderedSlice(collections []api.UserSubjectCollection, index int) []api.UserSubjectCollection {
	if index < 0 || index >= len(collections) {
		return collections
	}
	collection := collections[index]
	collections = append(collections[:index], collections[index+1:]...)
	return append([]api.UserSubjectCollection{collection}, collections...)
}
