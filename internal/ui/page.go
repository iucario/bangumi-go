package ui

import "github.com/rivo/tview"

// Page is a page in the TUI application that is a tview primitive.
type Page interface {
	tview.Primitive
	GetName() string
}
