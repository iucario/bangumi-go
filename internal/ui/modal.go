package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Modal is a centered message window used to inform the user or prompt them
// for an immediate decision. It needs to have at least one button (added via
// [Modal.AddButtons]) or it will never disappear.
//
// See https://github.com/rivo/tview/wiki/Modal for an example.
type Modal struct {
	*tview.Box

	// The frame embedded in the modal.
	frame *tview.Frame

	// The form embedded in the modal's frame.
	form *tview.Form

	// The message text (original, not word-wrapped).
	text string

	// The text color.
	textColor tcell.Color

	// The optional callback for when the user clicked one of the buttons. It
	// receives the index of the clicked button and the button's label.
	done func(buttonIndex int, buttonLabel string)
}

// NewModalForm returns a new form modal.
func NewModalForm(title string, form *tview.Form) *Modal {
	m := &Modal{
		Box: tview.NewBox().SetBorder(true).SetTitle(title).
			SetBackgroundColor(tview.Styles.ContrastBackgroundColor),
		text:      title,
		textColor: tview.Styles.PrimaryTextColor,
	}
	m.form = form
	m.form.SetBackgroundColor(tview.Styles.PrimitiveBackgroundColor)
	m.form.SetBorderPadding(1, 1, 1, 1)
	m.form.SetCancelFunc(func() {
		if m.done != nil {
			m.done(-1, "")
		}
	})
	frame := tview.NewFrame(m.form)
	frame.SetBorders(0, 0, 1, 0, 0, 0)
	frame.Box.SetBorderPadding(1, 1, 1, 1)
	frame.Box.SetBackgroundColor(tview.Styles.ContrastBackgroundColor)
	m.frame = frame

	return m
}

// Draw draws this primitive onto the screen.
func (m *Modal) Draw(screen tcell.Screen) {
	screenWidth, screenHeight := screen.Size()

	// Calculate modal size based on form content
	formWidth := 60                               // Minimum width for the form
	formHeight := m.form.GetFormItemCount()*3 + 4 // Rough estimate of height needed

	// Set the modal's position and size
	width := formWidth + 4   // Add padding
	height := formHeight + 4 // Add padding
	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2

	// Set the modal's rect
	m.SetRect(x, y, width, height)

	// Draw the background and border
	m.Box.DrawForSubclass(screen, m)

	// Get inner rect for the form
	innerX, innerY, innerWidth, innerHeight := m.GetInnerRect()

	// Position and draw the frame containing the form
	m.frame.SetRect(innerX, innerY, innerWidth, innerHeight)
	m.frame.Draw(screen)

	// Draw the form itself
	m.form.SetRect(innerX, innerY, innerWidth, innerHeight)
	m.form.Draw(screen)
}

// SetDoneFunc sets a handler which is called when one of the buttons was
// pressed. It receives the index of the button as well as its label text. The
// handler is also called when the user presses the Escape key. The index will
// then be negative and the label text an empty string.
func (m *Modal) SetDoneFunc(handler func(buttonIndex int, buttonLabel string)) *Modal {
	m.done = handler
	return m
}

// SetFocus shifts the focus to the button with the given index.
func (m *Modal) SetFocus(index int) *Modal {
	m.form.SetFocus(index)
	return m
}

// Focus is called when this primitive receives focus.
func (m *Modal) Focus(delegate func(p tview.Primitive)) {
	delegate(m.form)
}

// HasFocus returns whether or not this primitive has focus.
func (m *Modal) HasFocus() bool {
	return m.form.HasFocus()
}

// MouseHandler returns the mouse handler for this primitive.
func (m *Modal) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return m.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		// Pass mouse events on to the form.
		consumed, capture = m.form.MouseHandler()(action, event, setFocus)
		if !consumed && action == tview.MouseLeftDown && m.InRect(event.Position()) {
			setFocus(m)
			consumed = true
		}
		return
	})
}

// InputHandler returns the handler for this primitive.
func (m *Modal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if m.frame.HasFocus() {
			if handler := m.frame.InputHandler(); handler != nil {
				handler(event, setFocus)
				return
			}
		}
	})
}
