package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd/search"
	"github.com/iucario/bangumi-go/internal/ui"
	"github.com/rivo/tview"
)

type SearchPage struct {
	*tview.Grid
	client         *api.HTTPClient
	app            *App
	table          *tview.Table
	searchInput    *tview.InputField
	tagInput       *tview.InputField
	results        []api.Subject
	typeCheckboxes []*tview.Checkbox
	statusBar      *ui.StatusBar
}

func NewSearchPage(app *App) *SearchPage {
	typeLabels := api.S_TYPE_ALL
	typeCheckboxes := make([]*tview.Checkbox, len(typeLabels))
	for i, label := range typeLabels {
		cb := tview.NewCheckbox().SetLabel(label.CN())
		typeCheckboxes[i] = cb
	}
	search := &SearchPage{
		Grid:           tview.NewGrid(),
		client:         app.User.Client.HTTPClient,
		app:            app,
		table:          tview.NewTable(),
		searchInput:    tview.NewInputField().SetLabel("关键词: ").SetFieldWidth(40),
		tagInput:       tview.NewInputField().SetLabel("标签: ").SetFieldWidth(40),
		typeCheckboxes: typeCheckboxes,
		statusBar:      ui.NewStatusBar(),
	}
	search.fetchData()
	search.render()
	search.setKeyBindings()
	app.SetFocus(search.searchInput)
	return search
}

// search fetches data and render the ui accordingly
func (p *SearchPage) search() {
	p.fetchData()
	p.statusBar.SetMessage(fmt.Sprintf("Found %d results", len(p.results)), "success")
	p.render()
	// Focus the results table after rendering
	p.app.SetFocus(p.table)
}

func (p *SearchPage) fetchData() {
	p.statusBar.SetMessage("Searching...", "info")
	filter := api.Filter{
		MetaTags: []string{},
		Tag:      []string{},
		AirDate:  []string{},
		Rating:   []string{},
		Rank:     []string{},
		NSFW:     false,
	}
	keyword := p.searchInput.GetText()
	tags := p.tagInput.GetText()
	selectedTypes := p.CheckedTypes()
	if keyword == "" && tags == "" && len(selectedTypes) == 0 {
		p.statusBar.SetMessage("Please enter a search query or tags", "warning")
		return
	}
	if tags != "" {
		filter.Tag = strings.Split(tags, " ")
	}
	filter.Type = selectedTypes
	payload := api.Payload{
		Keyword: keyword,
		Sort:    api.MATCH,
		Filter:  filter,
	}
	pagesize := 20
	offset := 0
	result, err := search.Search(p.client, payload, pagesize, offset)
	if err != nil {
		p.statusBar.SetMessage(fmt.Sprintf("Error searching: %v error", err.Error()), "error")
		return
	}
	p.results = result.Data
}

func (p *SearchPage) render() {
	p.searchInput.SetLabel("Search: ").SetFieldWidth(40)
	p.tagInput.SetLabel("Tags: ").SetFieldWidth(40)
	p.Grid.SetRows(2, 2, 2, 0, 1)
	p.Grid.SetColumns(0)
	// Create typeFlex locally
	typeFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	for _, cb := range p.typeCheckboxes {
		typeFlex.AddItem(cb, 14, 0, false)
	}
	p.Grid.AddItem(p.searchInput, 0, 0, 1, 1, 0, 0, true)
	p.Grid.AddItem(p.tagInput, 1, 0, 1, 1, 0, 0, false)
	p.Grid.AddItem(typeFlex, 2, 0, 1, 1, 0, 0, false)
	p.Grid.AddItem(p.table, 3, 0, 1, 1, 0, 0, false)
	p.Grid.AddItem(p.statusBar, 4, 0, 1, 1, 0, 0, false)

	// Set up the results table
	p.table.SetBorders(false)
	p.table.SetSelectable(true, false)

	// Bind select function to open subject page
	p.table.SetSelectedFunc(func(row, column int) {
		if row > 0 && row <= len(p.results) {
			subject := p.results[row-1]
			p.app.OpenSubjectPage(int(subject.ID), "search")
		}
	})

	// Add headers
	p.table.SetCell(0, 0, tview.NewTableCell("Title").SetTextColor(tcell.ColorYellow))
	p.table.SetCell(0, 1, tview.NewTableCell("Type").SetTextColor(tcell.ColorYellow))
	p.table.SetCell(0, 2, tview.NewTableCell("Tags").SetTextColor(tcell.ColorYellow))
	p.table.SetCell(0, 3, tview.NewTableCell("Score").SetTextColor(tcell.ColorYellow))

	// Add results to table
	for i, subject := range p.results {
		row := i + 1
		// Set the subject ID as the reference for the first cell
		p.table.SetCell(row, 0, tview.NewTableCell(subject.GetName()).SetReference(int(subject.ID)))
		p.table.SetCell(row, 1, tview.NewTableCell(api.SubjectType(subject.Type).CN()))
		// Format score
		scoreStr := fmt.Sprintf("%.1f", subject.Rating.Score)
		p.table.SetCell(row, 2, tview.NewTableCell(subject.GetTags(5)))
		p.table.SetCell(row, 3, tview.NewTableCell(scoreStr))
	}
}

func (p *SearchPage) makeInputCapture(next, prev tview.Primitive, enterTriggersFetch bool) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if enterTriggersFetch {
				p.search()
				return nil
			}
		case tcell.KeyTAB:
			if next != nil {
				p.app.SetFocus(next)
				return nil
			}
		case tcell.KeyBacktab:
			if prev != nil {
				p.app.SetFocus(prev)
				return nil
			}
		case tcell.KeyEsc:
			p.app.SetFocus(p.table)
			return nil
		}
		return event
	}
}

func (p *SearchPage) setKeyBindings() {
	p.searchInput.SetInputCapture(p.makeInputCapture(p.tagInput, nil, true))
	p.tagInput.SetInputCapture(p.makeInputCapture(p.typeCheckboxes[0], p.searchInput, true))
	for idx, cb := range p.typeCheckboxes {
		cb.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEnter:
				p.search()
				return nil
			case tcell.KeyTAB:
				if idx < len(p.typeCheckboxes)-1 {
					p.app.SetFocus(p.typeCheckboxes[idx+1])
				} else {
					p.app.SetFocus(p.table)
				}
				return nil
			case tcell.KeyBacktab:
				if idx > 0 {
					p.app.SetFocus(p.typeCheckboxes[idx-1])
				} else {
					p.app.SetFocus(p.tagInput)
				}
				return nil
			case tcell.KeyLeft, tcell.KeyUp:
				if idx > 0 {
					p.app.SetFocus(p.typeCheckboxes[idx-1])
				}
				return nil
			case tcell.KeyRight, tcell.KeyDown:
				if idx < len(p.typeCheckboxes)-1 {
					p.app.SetFocus(p.typeCheckboxes[idx+1])
				}
				return nil
			case tcell.KeyRune:
				if event.Rune() == ' ' {
					cb.SetChecked(!cb.IsChecked())
					return nil
				}
			}
			return event
		})
	}
	p.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyBacktab {
			p.app.SetFocus(p.typeCheckboxes[len(p.typeCheckboxes)-1])
			return nil
		}
		if event.Key() == tcell.KeyRune && event.Rune() == ' ' {
			if cell := p.table.GetCell(p.table.GetSelection()); cell != nil && cell.GetReference() != nil {
				if p.app != nil {
					if subjectID, ok := cell.GetReference().(int); ok {
						p.app.OpenSubjectPage(subjectID, "calendar")
						return nil
					}
				}
			}
			return nil
		}
		if event.Key() == tcell.KeyRune {
			p.app.handlePageSwitch(event.Rune())
		}
		return event
	})
}

// CheckedTypes assumes the checkboxes are ordered the same as api.S_TYPE_ALL
// and returns selected subject types from indices.
func (p *SearchPage) CheckedTypes() []api.SubjectType {
	var selectedTypes []api.SubjectType
	for i, option := range p.typeCheckboxes {
		if option.IsChecked() {
			selectedTypes = append(selectedTypes, api.S_TYPE_ALL[i])
		}
	}
	return selectedTypes
}

func (p *SearchPage) GetName() string {
	return "search"
}
