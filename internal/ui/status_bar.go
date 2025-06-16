package ui

import (
	"github.com/rivo/tview"
)

// StatusBar represents the status bar at the bottom of the screen
type StatusBar struct {
	*tview.TextView
	message string
}

// NewStatusBar creates a new status bar
func NewStatusBar() *StatusBar {
	bar := &StatusBar{
		TextView: tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter),
	}
	return bar
}

// SetMessage displays a message in the status bar
// color should be one of: "info" (blue), "success" (green), "warning" (yellow), "error" (red)
func (s *StatusBar) SetMessage(message string, messageType string) {
	var coloredMessage string
	switch messageType {
	case "success":
		coloredMessage = Green(message)
	case "warning":
		coloredMessage = Yellow(message)
	case "error":
		coloredMessage = Red(message)
	default: // info
		coloredMessage = message
	}
	s.message = coloredMessage
	s.SetText(s.message)
}

// Clear clears the status bar message
func (s *StatusBar) Clear() {
	s.message = ""
	s.SetText("")
}
