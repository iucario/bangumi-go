package tui

import (
	"log/slog"
	"slices"

	"github.com/iucario/bangumi-go/api"
	"github.com/rivo/tview"

	"github.com/iucario/bangumi-go/internal/ui"
)

var PAGE_SIZE = 20

// App controls the whole UI
type App struct {
	*tview.Application
	Pages       *tview.Pages
	User        *api.User
	pageHistory []string // stack of page names for back navigation
}

func NewApp(user *api.User) *App {
	slog.Debug("New App", "User:", user)
	// Override the default styles
	tview.Styles = ui.Styles
	return &App{
		Application: tview.NewApplication(),
		Pages:       tview.NewPages(),
		User:        user,
	}
}

// Run starts the TUI application with watching list and sets up the main pages.
func (a *App) Run() error {
	// Add separate pages for each collection type
	// a.OpenSubjectPage(430699) // For testing
	a.Pages.AddPage("watching", NewCollectionPage(a, api.Watching), true, false)
	a.Pages.AddPage("wish", NewCollectionPage(a, api.Wish), true, false)
	a.Pages.AddPage("done", NewCollectionPage(a, api.Done), true, false)
	a.Pages.AddPage("stashed", NewCollectionPage(a, api.OnHold), true, false)
	a.Pages.AddPage("dropped", NewCollectionPage(a, api.Dropped), true, false)
	a.Pages.AddPage("calendar", NewCalendarPage(a), true, false)
	a.Pages.AddPage("help", NewHelpPage(a), true, false)
	a.Pages.SwitchToPage("watching")

	if err := a.Application.SetRoot(a.Pages, true).SetFocus(a.Pages).Run(); err != nil {
		panic(err)
	}
	return nil
}

// GoHome switchs app to page "watching"
func (a *App) GoHome() {
	a.Pages.SwitchToPage("watching")
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

// OpenSubjectPage pushes the current page to history and opens a subject page
func (a *App) OpenSubjectPage(subjectID int, page string) {
	a.PushPage(page)
	a.Pages.AddAndSwitchToPage("subject", NewSubjectPage(a, subjectID), true)
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
	case '6':
		a.Pages.SwitchToPage("calendar")
	case 'Q':
		a.Stop()
	case '?':
		a.Pages.SwitchToPage("help")
	}
}
