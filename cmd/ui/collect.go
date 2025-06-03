package tui

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/internal/ui"
	"github.com/iucario/bangumi-go/util"
	"github.com/rivo/tview"
)

var STATUS_LIST = []string{"wish", "done", "watching", "stashed", "dropped"}

// NewEditModal creates a modal for a collected subject
func NewEditModal(a *App, collection api.UserSubjectCollection, onSave func(*api.UserSubjectCollection)) *ui.Modal {
	closeFn := func() {
		a.Pages.RemovePage("collect")
		a.SetFocus(a.Pages) // Restore focus to the main page
	}
	form := createForm(collection, closeFn, onSave)
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
			slog.Debug("modal closed")
			closeFn()
		} else {
			slog.Debug(fmt.Sprintf("button %d clicked", buttonIndex))
			slog.Debug(fmt.Sprintf("button %s clicked", buttonLabel))
		}
	})

	return modal
}

// createForm creates a form for editing the collection.
func createForm(collection api.UserSubjectCollection, closeFn func(), onSave func(*api.UserSubjectCollection)) *tview.Form {
	status := util.IndexOfString(STATUS_LIST, collection.GetStatus().String())
	initTags := collection.GetTags()

	form := tview.NewForm()
	form.SetBorder(true).SetTitle(fmt.Sprintf("Edit Subject %d", collection.Subject.ID)).SetTitleAlign(tview.AlignLeft)
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
		onSave(&collection)
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
