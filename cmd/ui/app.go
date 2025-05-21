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
var PAGE_SIZE = 20

type App struct {
	*tview.Application
	Pages           *tview.Pages
	User            *api.User
	UserCollections map[api.CollectionStatus][]api.UserSubjectCollection
	listViews       map[api.CollectionStatus]*tview.List
	collectionTotal map[api.CollectionStatus]int
}

func NewApp(user *api.User) *App {
	slog.Debug("New App", "User:", user)
	// Override the default styles
	tview.Styles = ui.Styles
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
		listViews:       map[api.CollectionStatus]*tview.List{},
		collectionTotal: map[api.CollectionStatus]int{},
	}
}

// Run starts the TUI application with watching list and sets up the main pages.
func (a *App) Run() error {
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
	listView := tview.NewList()
	listView.SetBorder(true).SetTitle(fmt.Sprintf("List %s", collectionStatus)).SetTitleAlign(tview.AlignLeft)

	// Store list view reference for pagination
	a.listViews[collectionStatus] = listView

	options := list.UserListOptions{
		CollectionType: collectionStatus,
		Username:       a.User.Username,
		SubjectType:    "all",
		Limit:          PAGE_SIZE,
		Offset:         0,
	}
	userCollections, err := list.ListUserCollection(a.User.Client, options)
	if err != nil {
		slog.Error("Failed to fetch collections", "Error", err)
		return nil
	}
	a.UserCollections[collectionStatus] = userCollections.Data
	a.collectionTotal[collectionStatus] = int(userCollections.Total)
	collections := a.UserCollections[collectionStatus]

	// Update title to show total items
	listView.SetTitle(fmt.Sprintf("List %s (%d/%d)", collectionStatus, len(collections), userCollections.Total))

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
			case 'e':
				slog.Debug("edit")
				if len(collections) == 0 {
					slog.Warn("No collection to edit")
					return event
				}
				index := listView.GetCurrentItem()
				modal := a.NewEditModel(collections[index])
				a.Pages.AddPage("edit", modal, true, true) // Add modal as an overlay
				a.SetFocus(modal)                          // Set focus to the modal
			case 'R': // Refresh the list
				slog.Debug("refresh")
				a.ReloadCollection()
			case 'n': // Next page
				slog.Debug("next page")
				a.LoadPage(collectionStatus)
			case '1', '2', '3', '4', '5', 'q', 'Q':
				a.handlePageSwitch(event.Rune())
			}
		}
		return event
	})

	return collectionPage
}

func (a *App) NewHelpPage() *tview.TextView {
	text := `Welcome to Bangumi CLI UI
	Shortcuts:

	[General]
	1: Go to watching list
	2: Go to wish list
	3: Go to done list
	4: Go to stashed list
	5: Go to dropped list
	6: Go to Calendar
	7: Go to search
	0: Go to user info
	?: Show this help
	j/up: Move up
	k/down: Move down
	h/left: Switch to left
	l/right: Switch to right
	q/Q: Quit

	[Collection List]
	e: Edit collection
	shift + r: Refresh list
	n: Load next page
	`
	welcomePage := tview.NewTextView().SetText(text)

	welcomePage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			a.handlePageSwitch(event.Rune())
		}
		return event
	})
	return welcomePage
}

// LoadPage loads next page of a collection list
func (a *App) LoadPage(collectionStatus api.CollectionStatus) {
	size := len(a.UserCollections[collectionStatus])
	total := a.collectionTotal[collectionStatus]

	// Don't load more if we have all items
	if size >= total {
		slog.Info("No more items to load")
		return
	}

	options := list.UserListOptions{
		CollectionType: collectionStatus,
		Username:       a.User.Username,
		SubjectType:    "all",
		Limit:          PAGE_SIZE,
		Offset:         size,
	}
	c, err := list.ListUserCollection(a.User.Client, options)
	if err != nil {
		slog.Error("Failed to fetch collections", "Error", err)
		return
	}

	// Update collections and list view
	a.UserCollections[collectionStatus] = append(a.UserCollections[collectionStatus], c.Data...)
	listView := a.listViews[collectionStatus]
	if listView == nil {
		slog.Error("List view not found")
		return
	}

	// Add new items to list view
	for _, collection := range c.Data {
		name := collection.Subject.NameCn
		if name == "" {
			name = collection.Subject.Name
		}
		listView.AddItem(name, "", 0, nil)
	}

	// Update title to show progress
	listView.SetTitle(fmt.Sprintf("List %s (%d/%d)", collectionStatus, len(a.UserCollections[collectionStatus]), total))
}

// ReloadCollection recreates the user collection pages.
func (a *App) ReloadCollection() {
	currentPageName, _ := a.Pages.GetFrontPage()
	for _, status := range STATUS_LIST {
		a.Pages.RemovePage(status)
		a.Pages.AddPage(status, a.NewCollectionPage(api.CollectionStatus(status)), true, false)
	}
	a.Pages.SwitchToPage(currentPageName)
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

// newCollectionDetail creates a text view for the selected collection.
func newCollectionDetail(userCollection *api.UserSubjectCollection) *tview.TextView {
	subjectView := tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	subjectView.SetBorder(true).SetTitle("Subject Info").SetTitleAlign(tview.AlignLeft)
	subjectView.SetText(createCollectionText(userCollection))
	return subjectView
}

// Refactor colorToStr function to a utility function in the ui package
func colorToHex(color tcell.Color) string {
	r, g, b := color.RGB()
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// Update createCollectionText to use the new utility function
func createCollectionText(c *api.UserSubjectCollection) string {
	if c == nil {
		return "No data"
	}
	rate := fmt.Sprintf("%d", c.Rate)
	if c.Rate == 0 {
		rate = "Unknown"
	}
	totalEp := fmt.Sprintf("%d", c.Subject.Eps)
	if c.Subject.Eps != 0 {
		totalEp = "Unknown"
	}
	text := fmt.Sprintf("[%s]%s[-]\n%s\n\n", colorToHex(ui.Styles.SecondaryTextColor), c.Subject.NameCn, c.Subject.Name)
	text += fmt.Sprintf("%s\n", api.SubjectTypeRev[int(c.Subject.Type)])
	text += fmt.Sprintf("%s\n", c.Subject.ShortSummary)
	text += fmt.Sprintf("\nYour Tags: [%s]%s[-]\n", colorToHex(ui.Styles.TertiaryTextColor), c.GetTags())
	text += fmt.Sprintf("Your Rate: [%s]%s[-]\n", colorToHex(ui.Styles.TertiaryTextColor), rate)
	text += fmt.Sprintf("Episodes Watched: [%s]%d[-] of %s\n", colorToHex(ui.Styles.TertiaryTextColor), c.EpStatus, totalEp)
	if c.Subject.Type == 1 { // Book
		totalVol := "Unknown"
		if c.Subject.Volumes != 0 {
			totalVol = fmt.Sprintf("%d", c.Subject.Volumes)
		}
		text += fmt.Sprintf("Volumes Read: [%s]%d[-] of %s\n", colorToHex(ui.Styles.TertiaryTextColor), c.VolStatus, totalVol)
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
	form.AddInputField("Tags(Separate by spaces)", initTags, 0, nil, func(text string) {
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

func (a *App) handlePageSwitch(key rune) {
	switch key {
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
	case 'q', 'Q':
		a.Stop()
	}
}
