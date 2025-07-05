package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/iucario/bangumi-go/internal/ui"
	"github.com/rivo/tview"
)

type Help struct {
	*tview.TextView
	app *App
}

func NewHelpPage(app *App) *Help {
	text := `
    Welcome to Bangumi TUI
    <https://github.com/iucario/bangumi-go>

    Shortcuts:

    [%s]Pages[-]                               [%s]General[-]
    1: Go to watching list              e: Edit collection
    2: Go to wish list                  R: Refresh list
    3: Go to done list                  n: Load next page
    4: Go to stashed list               Space/Enter: View subject
    5: Go to dropped list               
    6: Go to calendar
    7: Go to search         
    0: Go to user info (not yet)
    ?: Show this help

    [%s]Navigation[-]
    j/up: Move up
    k/down: Move down
    h/left: Switch to left
    l/right: Switch to right
    Q: Quit
    q/Esc: Back
    `
	color := ui.ColorToHex(ui.Styles.TertiaryTextColor)
	colorText := fmt.Sprintf(text, color, color, color)
	view := tview.NewTextView().SetText(colorText).SetDynamicColors(true)
	view.SetWrap(false)

	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			app.handlePageSwitch(event.Rune())
		}
		return event
	})
	return &Help{
		TextView: view,
		app:      app,
	}
}

func (h *Help) GetName() string {
	return "help"
}
