package tui

import (
	"log/slog"
	"slices"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/iucario/bangumi-go/api"
	"github.com/rivo/tview"

	"github.com/iucario/bangumi-go/internal/ui"
)

var PAGES = []string{
	"watching",
	"wish",
	"done",
	"stashed",
	"dropped",
	"calendar",
	"help",
	"subject",
}

var MODALS = []string{
	"alert",
	"collect",
}

// App controls the whole UI
type App struct {
	*tview.Application
	Pages       *tview.Pages
	User        *api.User
	currentPage string
	pageHistory []string // stack of page names for back navigation
	statusBar   *ui.StatusBar
}

func NewApp(user *api.User) *App {
	// Override the default styles
	tview.Styles = ui.Styles
	return &App{
		Application: tview.NewApplication(),
		Pages:       tview.NewPages(),
		User:        user,
		statusBar:   ui.NewStatusBar(),
	}
}

// Run starts the TUI application with watching list and sets up the main pages.
func (a *App) Run() error {
	var wg sync.WaitGroup
	wg.Add(len(api.C_STATUS) + 2) // Collection pages + static pages

	// Create all pages concurrently using sync.Map for thread safety
	var collectionPages sync.Map
	pageCreation := func(page ui.Page) {
		defer wg.Done()
		name := page.GetName()
		a.Pages.AddPage(name, page, true, false)
		if slices.Contains(api.C_STATUS, api.CollectionStatus(name)) {
			if cp, ok := page.(*CollectionPage); ok {
				collectionPages.Store(name, cp)
			} else {
				slog.Error("Failed to cast page to CollectionPage", "Name", name)
			}
		}
	}

	// Create collection + other pages concurrently
	for _, status := range api.C_STATUS {
		go func(status api.CollectionStatus) {
			page := NewCollectionPage(a, status)
			pageCreation(page)
		}(status)
	}
	go func() {
		pageCreation(NewCalendarPage(a))
	}()
	go func() {
		pageCreation(NewHelpPage(a))
	}()

	// Wait for all pages to be created
	wg.Wait()

	// Switch to the initial page
	a.Goto("calendar")
	a.PushPage("calendar")

	// Start the application
	container := tview.NewGrid()
	container.SetRows(0, 1)
	container.SetBorder(false)
	container.SetBorders(false)
	container.AddItem(a.Pages, 0, 0, 1, 1, 0, 0, true)
	container.AddItem(a.statusBar, 1, 0, 1, 1, 0, 0, false)

	// Set up global input capture to clear status bar on user interaction
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if a.statusBar != nil {
			a.statusBar.Clear()
		}
		return event
	})

	return a.Application.SetRoot(container, true).SetFocus(a.Pages).Run()
}

// GoHome switchs app to page "watching"
func (a *App) GoHome() {
	a.Goto("watching")
}

// Switch to a page and set the page to current page
func (a *App) Goto(page string) {
	if ok := slices.Contains(PAGES, page); !ok {
		slog.Error("Invalid page name", "Page", page)
		return
	}
	// Should push subject page to history if switching from there
	if a.currentPage == "subject" {
		a.PushPage(a.currentPage)
	}
	a.Pages.SwitchToPage(page)
	a.currentPage = page
}

// PushPage adds a page to the history stack
func (a *App) PushPage(name string) {
	a.pageHistory = append(a.pageHistory, name)
}

// PopPage removes and returns the last page from the history stack
func (a *App) PopPage() (string, bool) {
	if len(a.pageHistory) == 0 {
		return "", false
	}
	index := len(a.pageHistory) - 1
	name := a.pageHistory[index]
	a.pageHistory = slices.Delete(a.pageHistory, index, index+1)
	return name, true
}

// Go back if there is a history, else do nothing
func (a *App) GoBack() {
	if prev, ok := a.PopPage(); ok {
		a.Goto(prev)
	}
}

// OpenSubjectPage pushes the current page to history and opens a subject page
func (a *App) OpenSubjectPage(subjectID int, prevPage string) {
	a.PushPage(prevPage)
	page := NewSubjectPage(a, subjectID)
	if page == nil {
		slog.Error("opening subject page", "ID", subjectID)
		return
	}
	a.Pages.AddPage("subject", page, true, false)
	a.Goto("subject")
}

func (a *App) OpenHelpPage() {
	a.PushPage(a.currentPage)
	a.Goto("help")
}

// Notify shows a notification message in the status bar.
func (a *App) Notify(message string) {
	if a.statusBar != nil {
		a.statusBar.SetMessage(message, "info") // default to info style
	}
}

// NotifyWithStyle shows a notification message in the status bar with a specific style.
// style can be "info", "success", "warning", or "error"
func (a *App) NotifyWithStyle(message string, style string) {
	if a.statusBar != nil {
		a.statusBar.SetMessage(message, style)
	}
}

// Alert displays a modal pop-up with a message.
func (a *App) Alert(message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.Pages.RemovePage("alert")
		})
	modal.SetTitle("Alert").SetTitleColor(ui.Styles.GraphicsColor)
	a.Pages.AddPage("alert", modal, true, true)
}

func (a *App) handlePageSwitch(key rune) {
	switch key {
	case '1':
		a.Goto("watching")
	case '2':
		a.Goto("wish")
	case '3':
		a.Goto("done")
	case '4':
		a.Goto("stashed")
	case '5':
		a.Goto("dropped")
	case '6':
		a.Goto("calendar")
	case 'Q':
		a.Stop()
	case 'q', rune(tcell.KeyEsc):
		a.GoBack()
	case '?':
		a.OpenHelpPage()
	}
}
