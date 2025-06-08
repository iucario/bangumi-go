package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Help struct {
	*tview.TextView
	app *App
}

func NewHelpPage(app *App) *Help {
	text := `Welcome to Bangumi CLI UI
	Shortcuts:

	[General]
	1: Go to watching list
	2: Go to wish list
	3: Go to done list
	4: Go to stashed list
	5: Go to dropped list
	6: Go to calendar
	7: Go to search (not yet)
	0: Go to user info (not yet)
	?: Show this help

	j/up: Move up
	k/down: Move down
	h/left: Switch to left
	l/right: Switch to right
	Q: Quit
	q: Back

	[Collection List]
	e: Edit collection
	R: Refresh list
	n: Load next page

	[Subject Page]
	e: Edit collection

	[Calendar]
	Enter: View subject
	`
	view := tview.NewTextView().SetText(text)

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
