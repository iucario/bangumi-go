package tui

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd/subject"
	"github.com/iucario/bangumi-go/internal/ui"
	"github.com/iucario/bangumi-go/util"
	"github.com/rivo/tview"
)

// NewCollectModal creates a modal for an uncollected subject
func NewCollectModal(a *App, s *api.Subject) *ui.Modal {
	collection := api.UserSubjectCollection{
		SubjectType: s.Type,
		Subject:     s.ToSlimSubject(),
		Type:        uint32(api.CollectionType[api.Watching]),
	}

	return NewEditModal(a, collection)
}

// NewEditModal creates a modal for a collected subject
func NewEditModal(a *App, collection api.UserSubjectCollection) *ui.Modal {
	closeFn := func() {
		a.Pages.RemovePage("collect")
		a.SetFocus(a.Pages) // Restore focus to the main page
	}
	form := createForm(collection, a, closeFn)
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
func createForm(collection api.UserSubjectCollection, a *App, closeFn func()) *tview.Form {
	// FIXME: should disable shortcuts when in form
	status := util.IndexOfString(STATUS_LIST, collection.GetStatus().String())
	initTags := collection.GetTags()
	prevStatus := collection.GetStatus()

	form := tview.NewForm()
	form.SetBorder(true).SetTitle("Edit Collection").SetTitleAlign(tview.AlignLeft)
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
		saveFn(a, collection, prevStatus)
		closeFn()
	})
	form.AddButton("Cancel", func() {
		slog.Info("cancel button clicked")
		closeFn()
	})
	return form
}

func saveFn(a *App, collection api.UserSubjectCollection, prevStatus api.CollectionStatus) {
	slog.Debug("save button clicked")
	slog.Debug("posting collection...")
	credential, err := api.GetCredential()
	if err != nil {
		slog.Error("login required")
		// TODO: display error messsage
	}
	err = subject.PostCollection(credential.AccessToken, int(collection.SubjectID), collection.GetStatus(),
		collection.Tags, collection.Comment, int(collection.Rate), collection.Private)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to post collection: %v", err))
	}
	subject.WatchToEpisode(a.User.Client, int(collection.SubjectID), int(collection.EpStatus))

	// Reorder the list and update the data in the watch list
	collections := a.UserCollections[prevStatus]
	updatedIndex := indexOfCollection(collections, collection.SubjectID)
	if updatedIndex != -1 {
		newCollections := reorderedSlice(collections, updatedIndex)
		collections = newCollections
		collections[0] = collection
		// Update the page
		// FIXME: not updating the other pages
		newPage := a.NewCollectionPage(prevStatus)
		a.Pages.RemovePage(prevStatus.String())
		a.Pages.AddPage(prevStatus.String(), newPage, true, false)
		a.Pages.SwitchToPage(prevStatus.String())
		a.SetFocus(newPage)
	}
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
