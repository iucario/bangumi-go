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

type CollectModal struct {
	*ui.Modal
	app *App // Reference to the main app for closing the modal
}

func NewCollectModal(a *App, collection api.UserSubjectCollection, onSave func(*api.UserSubjectCollection) error) *CollectModal {
	modal := &CollectModal{
		Modal: nil,
		app:   a,
	}
	form := modal.createForm(collection, onSave)
	modal.Modal = ui.NewModalForm("Edit Collection", form)

	// Set input capture at the form level to catch Esc
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			modal.Close()
			return nil // Prevent event from propagating
		}
		return event
	})

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex < 0 {
			modal.Close()
		}
	})

	return modal
}

func (m *CollectModal) Close() {
	m.app.Pages.RemovePage("collect")
	m.app.SetFocus(m.app.Pages) // Restore focus to the main page.
}

// createForm creates a form for editing the collection.
func (m *CollectModal) createForm(collection api.UserSubjectCollection, onSave func(*api.UserSubjectCollection) error) *tview.Form {
	initTags := collection.GetTags()

	form := tview.NewForm()
	form.SetBorder(true).SetTitle(fmt.Sprintf("Edit Subject %d", collection.Subject.ID)).SetTitleAlign(tview.AlignLeft)
	form.AddInputField("Episodes watched", util.Uint32ToString(collection.EpStatus), 5, nil, func(text string) {
		if text == "" {
			return
		}
		epStatus, err := strconv.Atoi(text)
		if err != nil {
			slog.Error(fmt.Sprintf("invalid episode number %s", text))
			// FIXME: alert error
		}
		collection.EpStatus = uint32(epStatus)
	})

	statusIndex := util.IndexOfString(STATUS_LIST, collection.GetStatus().String())
	if statusIndex == -1 {
		statusIndex = 2 // Default to watching if not set
	}
	form.AddDropDown("Status", STATUS_LIST, statusIndex, func(option string, optionIndex int) {
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
			rate = 0 // Default to 0 if invalid
		}
		rate = max(0, min(10, rate))
		collection.Rate = uint32(rate)
	})
	form.AddTextArea("Comment", collection.Comment, 0, 3, 200, func(text string) {
		collection.Comment = text
	})
	form.AddCheckbox("Private", collection.Private, func(checked bool) {
		collection.Private = checked
	})
	form.AddButton("Save", func() {
		err := onSave(&collection)
		if err != nil {
			m.app.Alert(fmt.Sprintf("Failed to save collection: %v", err))
		}
		m.Close()
	})
	form.AddButton("Cancel", func() {
		m.Close()
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
