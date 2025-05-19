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
	a.Pages.AddAndSwitchToPage("watching", a.NewCollectionPage(api.Watching), true)
	a.Pages.AddPage("wish", a.NewCollectionPage(api.Wish), true, false)
	a.Pages.AddPage("done", a.NewCollectionPage(api.Done), true, false)
	a.Pages.AddPage("stashed", a.NewCollectionPage(api.OnHold), true, false)
	a.Pages.AddPage("dropped", a.NewCollectionPage(api.Dropped), true, false)
	a.Pages.AddPage("help", a.NewHelpPage(), true, false)

	if err := a.Application.SetRoot(a.Pages, true).SetFocus(a.Pages).Run(); err != nil {
		panic(err)
	}
	return nil
}

// NewCollectionPage creates a list with detail page for a specific collection type.
func (a *App) NewCollectionPage(collectionStatus api.CollectionStatus) *tview.Flex {
	mainTextStyle := tcell.StyleDefault.Foreground(ui.Styles.PrimaryTextColor).Background(ui.Styles.PrimitiveBackgroundColor)
	listView := tview.NewList().SetMainTextStyle(mainTextStyle)
	listView.SetBackgroundColor(ui.Styles.PrimitiveBackgroundColor)
	listView.SetBorder(true).SetTitle(fmt.Sprintf("List %s", collectionStatus)).SetTitleAlign(tview.AlignLeft)

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

	// TODO: move this to collection function
	for _, collection := range collections {
		name := collection.Subject.NameCn
		if name == "" {
			name = collection.Subject.Name
		}
		listView.AddItem(name, "", 0, nil)
	}

	// Initialize the first item's detail view
	detailView := newCollectionDetail(nil)
	if len(collections) > 0 {
		detailView = newCollectionDetail(&collections[0])
	}
	detailView.SetScrollable(true)

	// Update subject info when an item is selected
	listView.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(collections) {
			collection := collections[index]
			slog.Debug(fmt.Sprintf("Selected %s", collection.Subject.Name))
			detailView.SetText(createCollectionText(&collection))
		}
	})

	// The flex layout for the collection page
	collectionPage := tview.NewFlex().
		AddItem(listView, 0, 2, true).
		AddItem(detailView, 0, 3, false)
	collectionPage.SetBackgroundColor(ui.Styles.PrimitiveBackgroundColor)
	collectionPage.SetFullScreen(true).SetBorderPadding(0, 0, 0, 0)

	// Scroll with j/k keys
	detailView.SetInputCapture(handleScrollKeys(detailView))
	listView.SetInputCapture(handleScrollKeys(listView))

	// Update the input capture to use numeric keys for switching pages
	collectionPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
			a.SetFocus(listView)
		case tcell.KeyRight:
			a.SetFocus(detailView)
		case tcell.KeyRune:
			switch event.Rune() {
			case 'l':
				a.SetFocus(detailView)
			case 'h':
				a.SetFocus(listView)
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
				index := listView.GetCurrentItem()
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

// handleScrollKeys captures input events for the Box and handles 'j' and 'k' keys.
func handleScrollKeys(b tview.Primitive) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'j':
			b.InputHandler()(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone), nil)
		case 'k':
			b.InputHandler()(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone), nil)
		}
		return event
	}
}

func (a *App) NewHelpPage() *tview.TextView {
	text := `Welcome to Bangumi CLI UI
	Shortcuts:

	[Navigation]
	1: Go to watching list
	2: Go to wish list
	3: Go to done list
	4: Go to stashed list
	5: Go to dropped list
	?: Show this help
	j/up: Move up
	k/down: Move down
	h/left: Switch to left
	l/right: Switch to right

	[Collection List]
	e: Edit collection
	`
	welcomePage := tview.NewTextView().SetText(text)

	welcomePage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
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
	if c == nil {
		return "No data"
	}
	colorToStr := func(color tcell.Color) string {
		r, g, b := color.RGB()
		return fmt.Sprintf("#%02X%02X%02X", r, g, b)
	}

	text := fmt.Sprintf("[%s]%s[-]\n%s\n\n", colorToStr(ui.Styles.SecondaryTextColor), c.Subject.NameCn, c.Subject.Name)
	text += fmt.Sprintf("%s\n", api.SubjectTypeRev[int(c.Subject.Type)])
	text += fmt.Sprintf("%s\n", c.Subject.ShortSummary)
	text += fmt.Sprintf("\nYour Tags: [%s]%s[-]\n", colorToStr(ui.Styles.TertiaryTextColor), c.GetTags())
	rate := "Unknown"
	if c.Rate != 0 {
		rate = fmt.Sprintf("%d", c.Rate)
	}
	text += fmt.Sprintf("Your Rate: [%s]%s[-]\n", colorToStr(ui.Styles.TertiaryTextColor), rate)
	totalEp := "Unknown"
	if c.Subject.Eps != 0 {
		totalEp = fmt.Sprintf("%d", c.Subject.Eps)
	}
	text += fmt.Sprintf("Episodes Watched: [%s]%d[-] of %s\n", colorToStr(ui.Styles.TertiaryTextColor), c.EpStatus, totalEp)
	if c.Subject.Type == 1 { // Book
		totalVol := "Unknown"
		if c.Subject.Volumes != 0 {
			totalVol = fmt.Sprintf("%d", c.Subject.Volumes)
		}
		text += fmt.Sprintf("Volumes Read: [%s]%d[-] of %s\n", colorToStr(ui.Styles.TertiaryTextColor), c.VolStatus, totalVol)
	}
	text += "\n---------------------------------------\n\n"
	text += fmt.Sprintf("On Air Date: %s\n", c.Subject.Date)
	text += fmt.Sprintf("User Score: %.1f\n", c.Subject.Score)
	text += fmt.Sprintf("Rank: %d\n", c.Subject.Rank)
	text += fmt.Sprintf("User Tags: %s\n", c.GetAllTags())
	text += fmt.Sprintf("Marked By: %d users\n", c.Subject.CollectionTotal)
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
			newPage := a.NewCollectionPage(collectionStatus)
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
