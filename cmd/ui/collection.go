package tui

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/gdamore/tcell/v2"
	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd/list"
	"github.com/iucario/bangumi-go/cmd/subject"
	"github.com/iucario/bangumi-go/internal/ui"
	"github.com/rivo/tview"
)

var PAGE_SIZE = 20

type CollectionPage struct {
	*tview.Flex
	Name             string
	CollectionStatus api.CollectionStatus
	Collections      []api.UserSubjectCollection
	Total            int
	app              *App
	ListView         *tview.List
	DetailView       *tview.TextView
	CurrentSubject   int // Subject ID in selection
}

// NewCollectionPage creates a list with detail page for a specific collection type.
func NewCollectionPage(a *App, collectionStatus api.CollectionStatus) *CollectionPage {
	userCollections, err := list.ListUserCollection(a.User.Client, list.UserListOptions{
		CollectionType: collectionStatus,
		Username:       a.User.Username,
		SubjectType:    "all", // TODO: add filter feature
		Limit:          PAGE_SIZE,
		Offset:         0,
	})
	if err != nil {
		slog.Error("Failed to fetch collections", "Error", err)
		return nil
	}
	currentSubject := 0
	if len(userCollections.Data) > 0 {
		currentSubject = int(userCollections.Data[0].Subject.ID)
	}
	collectionPage := &CollectionPage{
		Flex:             tview.NewFlex(),
		app:              a,
		Name:             collectionStatus.String(),
		CollectionStatus: collectionStatus,
		Collections:      userCollections.Data,
		Total:            int(userCollections.Total),
		ListView:         nil,
		DetailView:       nil,
		CurrentSubject:   currentSubject,
	}
	collectionPage.render()
	collectionPage.setKeyBindings()
	return collectionPage
}

func (c *CollectionPage) GetName() string {
	return c.Name
}

// LoadData fetches and loads collection data asynchronously
func (c *CollectionPage) LoadData() {
	userCollections, err := list.ListUserCollection(c.app.User.Client, list.UserListOptions{
		CollectionType: c.CollectionStatus,
		Username:       c.app.User.Username,
		SubjectType:    "all",
		Limit:          PAGE_SIZE,
		Offset:         0,
	})
	if err != nil {
		slog.Error("Failed to fetch collections", "Error", err)
		return
	}

	c.Collections = userCollections.Data
	c.Total = int(userCollections.Total)
	if len(userCollections.Data) > 0 {
		c.CurrentSubject = int(userCollections.Data[0].Subject.ID)
	}

	c.renderListItems()
	c.renderDetail()
}

func (c *CollectionPage) render() {
	c.Clear()
	c.ListView = tview.NewList()
	c.ListView.SetBorder(true).SetTitleAlign(tview.AlignLeft)
	c.ListView.SetWrapAround(false)

	c.DetailView = newCollectionDetail(nil)
	c.DetailView.SetScrollable(true)
	c.renderListItems()
	c.renderDetail()

	// Scroll with j/k keys
	c.ListView.SetInputCapture(handleScrollKeys(c.ListView))
	c.DetailView.SetInputCapture(handleScrollKeys(c.DetailView))

	c.ListView.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(c.Collections) {
			slog.Debug(fmt.Sprintf("Selected %s", c.Collections[index].Subject.Name))
			c.CurrentSubject = int(c.Collections[index].Subject.ID)
			c.renderDetail()
		}
	})

	// Open Subject page on click(enter/space)
	c.ListView.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(c.Collections) {
			subID := int(c.Collections[index].Subject.ID)
			c.app.OpenSubjectPage(subID, c.Name)
		}
	})

	// Change border color on focus/blur
	c.ListView.SetFocusFunc(func() {
		c.ListView.SetBorderColor(ui.Styles.TitleColor) // Focused color
	})
	c.ListView.SetBlurFunc(func() {
		c.ListView.SetBorderColor(tcell.ColorGray) // Unfocused color
	})
	c.DetailView.SetFocusFunc(func() {
		c.DetailView.SetBorderColor(ui.Styles.TitleColor)
	})
	c.DetailView.SetBlurFunc(func() {
		c.DetailView.SetBorderColor(tcell.ColorGray)
	})

	c.Clear()
	c.Flex.AddItem(c.ListView, 0, 2, true).
		AddItem(c.DetailView, 0, 3, false)
	c.Flex.SetFullScreen(false).SetBorderPadding(0, 0, 0, 0)
}

func (c *CollectionPage) Refresh() {
	collections, err := list.ListUserCollection(c.app.User.Client, list.UserListOptions{
		CollectionType: c.CollectionStatus,
		Username:       c.app.User.Username,
		SubjectType:    "all", // TODO: add filter feature
		Limit:          PAGE_SIZE,
		Offset:         0,
	})
	if err != nil {
		slog.Error("Failed to refresh collections", "Error", err)
		return
	}

	c.Collections = collections.Data
	c.Total = int(collections.Total)
	if len(collections.Data) > 0 {
		c.CurrentSubject = int(collections.Data[0].Subject.ID)
	} else {
		c.CurrentSubject = 0
	}
	c.renderListItems()
	c.renderDetail()
}

// LoadNextPage loads next page of a collection list
func (c *CollectionPage) LoadNextPage() {
	size := len(c.Collections)

	// Don't load more if we have all items
	// TODO: can fetch API first to be assured
	if size >= c.Total {
		slog.Info("No more items to load")
		c.app.Notify("No more items")
		return
	}
	collections, err := list.ListUserCollection(c.app.User.Client, list.UserListOptions{
		CollectionType: c.CollectionStatus,
		Username:       c.app.User.Username,
		SubjectType:    "all",
		Limit:          PAGE_SIZE,
		Offset:         size,
	})
	if err != nil {
		slog.Error("Failed to fetch collections", "Error", err)
		return
	}

	// Update collections and list view
	c.Collections = append(c.Collections, collections.Data...)
	if c.ListView == nil {
		slog.Error("List view not found")
		return
	}

	// Add new items to list view
	for _, collection := range collections.Data {
		c.ListView.AddItem(collection.Name(), "", 0, nil)
	}

	// Update title to show progress
	c.ListView.SetTitle(fmt.Sprintf("List %s (%d/%d)", c.CollectionStatus, len(c.Collections), c.Total))
}

func (c *CollectionPage) setKeyBindings() {
	listView := c.ListView
	detailView := c.DetailView
	c.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
			c.app.SetFocus(listView)
		case tcell.KeyRight:
			c.app.SetFocus(detailView)
		case tcell.KeyRune:
			switch event.Rune() {
			case 'l':
				c.app.SetFocus(detailView)
			case 'h':
				c.app.SetFocus(listView)
			case 'e':
				slog.Debug("collect")
				if len(c.Collections) == 0 {
					slog.Warn("No collection to edit")
					return event
				}
				index := listView.GetCurrentItem()
				if index < 0 || index >= len(c.Collections) {
					slog.Warn("Invalid collection index for edit")
					return event
				}
				modal := NewCollectModal(c.app, &c.Collections[index], c.onSave)
				if modal != nil {
					c.app.Pages.AddPage("collect", modal, true, true)
					c.app.SetFocus(modal)
				} else {
					c.app.NotifyWithStyle("collection is nil", "error")
				}
			case 'R':
				c.Refresh()
			case 'n':
				c.LoadNextPage()
			default:
				c.app.handlePageSwitch(event.Rune())
			}
		}
		return event
	})
}

// Render the list view and detail view based on collection data
func (c *CollectionPage) renderListItems() {
	c.ListView.Clear()
	c.ListView.SetTitle(fmt.Sprintf("List %s (%d/%d)", c.CollectionStatus, len(c.Collections), c.Total))

	for _, collection := range c.Collections {
		c.ListView.AddItem(collection.Name(), "", 0, nil)
	}
}

// Render the detail view based on the current selection
func (c *CollectionPage) renderDetail() {
	currentIndex := indexOfCollection(c.Collections, uint32(c.CurrentSubject))
	if 0 <= currentIndex && currentIndex < len(c.Collections) {
		c.DetailView.SetText(createCollectionText(&c.Collections[currentIndex]))
	} else {
		c.DetailView.SetText("No data")
	}
}

// onSave post collection data if the data has changed. Watch to N episode if only changed.
func (c *CollectionPage) onSave(collection *api.UserSubjectCollection) error {
	// Find the original collection for comparison
	updatedIndex := indexOfCollection(c.Collections, collection.SubjectID)
	if updatedIndex < 0 {
		slog.Error("Collection not found in the list")
		return errors.New("collection not found in the list")
	}
	original := c.Collections[updatedIndex]

	// Collection info excluding episode/volume status
	err := subject.PatchCollection(c.app.User.Client, &original, collection)
	if err != nil {
		slog.Error("Saving collection", "Error", err)
		return err
	}
	// Episode/volume status update
	if EpisodeStatusChanged(&original, collection) {
		subject.WatchToEpisode(c.app.User.Client, int(collection.SubjectID), int(collection.EpStatus))
	}

	c.Collections = toFrontItem(c.Collections, updatedIndex)
	c.Collections[0] = *collection
	c.renderListItems()
	c.renderDetail()
	// TODO: update other collection page if needed
	// The updated collection may go to other collection page
	return nil
}

// Select a subject in the collection page by its ID.
// Updates the list and detail views accordingly, and sets the selection field.
func (c *CollectionPage) Select(subjectID int) {
	slog.Debug(fmt.Sprintf("Select subject %d in collection page", subjectID))
	index := indexOfCollection(c.Collections, uint32(subjectID))
	if index < 0 || index >= len(c.Collections) {
		slog.Error("Subject not found in the collection")
		return
	}
	c.ListView.SetCurrentItem(index)
	c.DetailView.SetText(createCollectionText(&c.Collections[index]))
	c.CurrentSubject = index
	c.app.SetFocus(c.ListView)
}

// newCollectionDetail creates a text view for the selected collection.
func newCollectionDetail(userCollection *api.UserSubjectCollection) *tview.TextView {
	subjectView := tview.NewTextView().SetDynamicColors(true).SetWrap(true)
	subjectView.SetBorder(true).SetTitle("Subject Info").SetTitleAlign(tview.AlignLeft)
	subjectView.SetText(createCollectionText(userCollection))
	return subjectView
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
	if c.Subject.Eps == 0 {
		totalEp = "Unknown"
	}
	text := ""
	if c.Subject.NameCn == "" {
		text += fmt.Sprintf("%s\n", ui.SecondaryText(c.Subject.Name))
	} else {
		text += fmt.Sprintf("%s\n%s\n", ui.SecondaryText(c.Subject.NameCn), c.Subject.Name)
	}
	text += fmt.Sprintf("%s\n\n", api.SubjectTypeRev[int(c.Subject.Type)])
	text += fmt.Sprintf("%s...\n", c.Subject.ShortSummary)
	text += fmt.Sprintf("\n你的标签: %s\n", ui.SecondaryText(c.GetTags()))
	text += fmt.Sprintf("你的评分: %s\n", ui.SecondaryText(rate))
	text += fmt.Sprintf("看到: %s of %s\n", ui.SecondaryText(fmt.Sprintf("%d", c.EpStatus)), totalEp)
	if c.Subject.Type == 1 { // Book
		totalVol := "Unknown"
		if c.Subject.Volumes != 0 {
			totalVol = fmt.Sprintf("%d", c.Subject.Volumes)
		}
		text += fmt.Sprintf("读到: %s of %s\n", ui.SecondaryText(fmt.Sprintf("%d", c.VolStatus)), totalVol)
	}
	text += fmt.Sprintf("隐私收藏: %v\n", c.Private)
	text += "\n---------------------------------------\n\n"
	text += fmt.Sprintf("放送日期: %s\n", c.Subject.Date)
	text += fmt.Sprintf("评分: %.1f\n", c.Subject.Score)
	text += fmt.Sprintf("排名: %d\n", c.Subject.Rank)
	text += fmt.Sprintf("用户标签: %s\n", c.GetAllTags())
	text += fmt.Sprintf("收藏人数: %d\n", c.Subject.CollectionTotal)
	return text
}

// Move the item at index to the front of the slice
func toFrontItem(collections []api.UserSubjectCollection, index int) []api.UserSubjectCollection {
	if index < 0 || index >= len(collections) {
		return collections
	}
	collection := collections[index]
	collections = slices.Delete(collections, index, index+1)
	newSlice := make([]api.UserSubjectCollection, 0, len(collections)+1)
	newSlice = append(newSlice, collection)
	newSlice = append(newSlice, collections...)
	return newSlice
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
