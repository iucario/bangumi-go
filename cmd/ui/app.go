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

	"slices"

	"github.com/iucario/bangumi-go/internal/ui"
)

var STATUS_LIST = []string{"wish", "done", "watching", "stashed", "dropped"}

type App struct {
	*tview.Application
	Pages           *tview.Pages
	User            *api.User
	UserCollections map[api.CollectionStatus][]api.UserSubjectCollection
}

func NewApp(user *api.User) *App {
	slog.Debug("New App", "User:", user)
	return &App{
		Application: tview.NewApplication(),
		Pages:       tview.NewPages(),
		User:        user,
		UserCollections: map[api.CollectionStatus][]api.UserSubjectCollection{
			api.Watching: {},
			api.Wish:     {},
			api.Done:     {},
			api.OnHold:   {},
			api.Dropped:  {},
		},
	}
}

// Run starts the TUI application with watching list and sets up the main pages.
func (a *App) Run() error {
	options := list.UserListOptions{
		SubjectType:    "all",
		Username:       a.User.Username,
		CollectionType: api.Watching,
		Limit:          20,
		Offset:         0,
	}
	userCollections, err := list.ListUserCollection(a.User.Client, options)
	if err != nil {
		return err
	}
	a.UserCollections[api.Watching] = userCollections.Data

	// Add separate pages for each collection type
	a.Pages.AddAndSwitchToPage("watching", a.NewListPage(api.Watching), true)
	a.Pages.AddPage("wish", a.NewListPage(api.Wish), true, false)
	a.Pages.AddPage("done", a.NewListPage(api.Done), true, false)
	a.Pages.AddPage("stashed", a.NewListPage(api.OnHold), true, false)
	a.Pages.AddPage("dropped", a.NewListPage(api.Dropped), true, false)
	a.Pages.AddPage("help", a.NewHelpPage(), true, false)

	if err := a.Application.SetRoot(a.Pages, true).SetFocus(a.Pages).Run(); err != nil {
		panic(err)
	}
	return nil
}

// NewListPage creates a list with detail page for a specific collection type.
func (a *App) NewListPage(collectionStatus api.CollectionStatus) *tview.Flex {
	mainTextStyle := tcell.StyleDefault.Foreground(ui.Styles.PrimaryTextColor).Background(ui.Styles.PrimitiveBackgroundColor)
	collectionList := tview.NewList().SetMainTextStyle(mainTextStyle)
	collectionList.SetBackgroundColor(ui.Styles.PrimitiveBackgroundColor)
	collectionList.SetBorder(true).SetTitle(fmt.Sprintf("List %s", collectionStatus)).SetTitleAlign(tview.AlignLeft)

	options := list.UserListOptions{
		CollectionType: collectionStatus,
		Username:       a.User.Username,
		SubjectType:    "all",
		Limit:          20,
		Offset:         0,
	}

	userCollections, err := list.ListUserCollection(a.User.Client, options)
	if err != nil {
		slog.Error("Failed to fetch collections", "Error", err)
		return nil
	}
	a.UserCollections[collectionStatus] = userCollections.Data
	collections := a.UserCollections[collectionStatus]

	for _, collection := range collections {
		name := collection.Subject.NameCn
		if name == "" {
			name = collection.Subject.Name
		}
		collectionList.AddItem(name, "", 0, nil)
	}

	collectionView := newCollectionDetail(&collections[0])

	pages := tview.NewPages()
	pages.AddAndSwitchToPage("view", collectionView, true)

	// Update subject info when an item is selected
	collectionList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(collections) {
			collection := collections[index]
			slog.Debug(fmt.Sprintf("Selected %s", collection.Subject.Name))
			collectionView.SetText(createCollectionText(&collection))
		}
	})

	// The flex layout for the collection page
	collectionPage := tview.NewFlex().
		AddItem(collectionList, 0, 2, true).
		AddItem(pages, 0, 3, false)
	collectionPage.SetBackgroundColor(ui.Styles.PrimitiveBackgroundColor)
	collectionPage.SetFullScreen(true).SetBorderPadding(0, 0, 0, 0)

	// Update the input capture to use numeric keys for switching pages
	collectionPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				collectionList.SetCurrentItem((collectionList.GetCurrentItem() + 1) % len(collections))
			case 'k':
				collectionList.SetCurrentItem((collectionList.GetCurrentItem() - 1) % len(collections))
			case '?':
				a.Pages.SwitchToPage("help")
			case '1':
				a.Pages.SwitchToPage("watching")
			case '2':
				a.Pages.SwitchToPage("wish")
			case '3':
				a.Pages.SwitchToPage("done")
			case '4':
				a.Pages.SwitchToPage("stashed")
			case '5':
				a.Pages.SwitchToPage("dropped")
			case 'e':
				slog.Debug("edit")
				index := collectionList.GetCurrentItem()
				modal := a.NewEditModel(collections[index])
				a.Pages.AddPage("edit", modal, true, true) // Add modal as an overlay
				a.SetFocus(modal)                          // Set focus to the modal
			case 'r': // Refresh the list
				slog.Debug("refresh")
				// TODO: implement refresh logic
			}
		}
		return event
	})

	return collectionPage
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

// newCollectionDetail creates a text view for the selected collection.
func newCollectionDetail(userCollection *api.UserSubjectCollection) *tview.TextView {
	subjectView := tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	subjectView.SetBackgroundColor(ui.Styles.PrimitiveBackgroundColor)
	subjectView.SetBorder(true).SetTitle("Subject Info").SetTitleAlign(tview.AlignLeft)
	subjectView.SetText(createCollectionText(userCollection))
	return subjectView
}

// createCollectionText generates the text to display details of a collection.
func createCollectionText(c *api.UserSubjectCollection) string {
	text := fmt.Sprintf("[yellow]%s[-]\n%s\n\n%s\n", c.Subject.NameCn, c.Subject.Name, c.Subject.ShortSummary)
	tags := strings.Join(c.Tags, ", ")
	text += fmt.Sprintf("\nYour Tags: [green]%s[-]\n", tags)
	if c.Rate == 0 {
		text += "Your Rate: [blue]N/A[-]\n"
	} else {
		text += fmt.Sprintf("Your Rate: [blue]%d[-]\n", c.Rate)
	}
	totalEp := "Unknown"
	if c.Subject.Eps != 0 {
		totalEp = fmt.Sprintf("%d", c.Subject.Eps)
	}
	text += fmt.Sprintf("Episodes Watched: %d of %s\n", c.EpStatus, totalEp)
	if c.Subject.Type == 1 { // Book
		totalVol := "Unknown"
		if c.Subject.Volumes != 0 {
			totalVol = fmt.Sprintf("%d", c.Subject.Volumes)
		}
		text += fmt.Sprintf("Volumes Read: %d of %s\n", c.VolStatus, totalVol)
	}
	text += fmt.Sprintf("On Aired: %s\n", c.Subject.Date)
	text += fmt.Sprintf("User Score: %.1f\n", c.Subject.Score)

	return text
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

// createForm creates a form for editing the collection.
func createForm(collection api.UserSubjectCollection, a *App, closeFn func()) *tview.Form {
	// FIXME: should disable shortcuts when in form
	status := util.IndexOfString(STATUS_LIST, collection.GetStatus().String())
	initTags := collection.GetTags()
	collectionStatus := collection.GetStatus()

	form := tview.NewForm()
	form.SetBorder(true).SetTitle("Edit Collection").SetTitleAlign(tview.AlignLeft)
	form.SetFieldBackgroundColor(ui.Styles.ContrastSecondaryTextColor)
	form.AddInputField("Episodes watched", util.Uint32ToString(collection.EpStatus), 5, nil, func(text string) {
		epStatus, err := strconv.Atoi(text)
		if err != nil {
			slog.Error(fmt.Sprintf("invalid episode status %s", text))
		}
		collection.EpStatus = uint32(epStatus)
	})

	form.AddDropDown("Status", STATUS_LIST, status, func(option string, optionIndex int) {
		slog.Debug(fmt.Sprintf("selected %s", option))
		collection.SetStatus(api.CollectionStatus(option))
	})
	form.AddInputField("Tags", initTags, 0, nil, func(text string) {
		// TODO: validate tags
		collection.Tags = strings.Split(text, " ")
	})
	form.AddInputField("Rate", util.Uint32ToString(collection.Rate), 3, nil, func(text string) {
		rate, err := strconv.Atoi(text)
		if err != nil {
			slog.Error(fmt.Sprintf("invalid rate %s. Must be in [0-10]", text))
		}
		rate = max(0, min(10, rate))
		collection.Rate = uint32(rate)
	})
	form.AddInputField("Comment", collection.Comment, 0, nil, func(text string) {
		collection.Comment = text
	})
	form.AddCheckbox("Private", collection.Private, func(checked bool) {
		collection.Private = checked
	})
	form.AddButton("Save", func() {
		slog.Debug("save button clicked")
		slog.Debug("posting collection...")
		credential, err := api.GetCredential()
		if err != nil {
			slog.Error("login required")
			// TODO: display error messsage
		}
		err = subject.PostCollection(credential.AccessToken, int(collection.SubjectID), collection.GetStatus(),
			collection.Tags, collection.Comment, int(collection.Rate), collection.Private)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to post collection: %v", err))
		}
		subject.WatchToEpisode(credential.AccessToken, int(collection.SubjectID), int(collection.EpStatus))

		// Reorder the list and update the data in the watch list
		collections := a.UserCollections[collectionStatus]
		updatedIndex := indexOfCollection(collections, collection.SubjectID)
		if updatedIndex != -1 {
			newCollections := reorderedSlice(collections, updatedIndex)
			collections = newCollections
			collections[0] = collection
			// Update the page
			// FIXME: not updating the other pages
			newPage := a.NewListPage(collectionStatus)
			a.Pages.RemovePage(collectionStatus.String())
			a.Pages.AddPage(collectionStatus.String(), newPage, true, false)
			a.Pages.SwitchToPage(collectionStatus.String())
			a.SetFocus(newPage)
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
	collections = slices.Delete(collections, index, index+1)
	return append([]api.UserSubjectCollection{collection}, collections...)
}
