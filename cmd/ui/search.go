package tui

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd/search"
	"github.com/iucario/bangumi-go/internal/ui"
	"github.com/rivo/tview"
)

type SearchPage struct {
	*tview.Grid
	client      *api.HTTPClient
	app         *App
	table       *tview.Table
	searchInput *tview.InputField
	tagInput    *tview.InputField
	results     []api.Subject
	statusBar   *ui.StatusBar
}

func NewSearchPage(app *App) *SearchPage {
	search := &SearchPage{
		Grid:        tview.NewGrid(),
		client:      app.User.Client.HTTPClient,
		app:         app,
		table:       tview.NewTable(),
		searchInput: tview.NewInputField().SetLabel("Keywords: ").SetFieldWidth(40),
		tagInput:    tview.NewInputField().SetLabel("Tags: ").SetFieldWidth(40),
		statusBar:   ui.NewStatusBar(),
	}
	search.fetchData()
	search.render()
	search.setKeyBindings()
	// Set default focus to the search input field
	app.SetFocus(search.searchInput)
	return search
}

func (p *SearchPage) fetchData() {
	p.statusBar.SetMessage("Searching...", "info")

	// Initialize a basic filter
	filter := api.Filter{
		MetaTags: []string{},
		Tag:      []string{},
		AirDate:  []string{},
		Rating:   []string{},
		Rank:     []string{},
		NSFW:     false,
	}

	// Get search query from input
	keyword := p.searchInput.GetText()
	tags := p.tagInput.GetText()
	if keyword == "" && tags == "" {
		p.statusBar.SetMessage("Please enter a search query or tags", "warning")
		return
	}
	if tags != "" {
		filter.Tag = strings.Split(tags, ",")
	}

	payload := api.Payload{
		Keyword: keyword,
		Sort:    api.MATCH, // or allow user to select
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
	p.statusBar.SetMessage(fmt.Sprintf("Found %d results", len(result.Data)), "success")
	p.render()
	slog.Debug("focusing", "after fetch data", p.app.GetFocus())
}

func (p *SearchPage) render() {
	// Set up the keyword and tag input
	p.searchInput.SetLabel("Search: ").SetFieldWidth(40)
	p.tagInput.SetLabel("Tags: ").SetFieldWidth(40)

	// Set up the grid layout (keyword, tag, table, status bar)
	p.Grid.SetRows(2, 2, 0, 1) // Input, tag, table, status bar
	p.Grid.SetColumns(0)       // Full width

	p.Grid.AddItem(p.searchInput, 0, 0, 1, 1, 0, 0, true)
	p.Grid.AddItem(p.tagInput, 1, 0, 1, 1, 0, 0, false)
	p.Grid.AddItem(p.table, 2, 0, 1, 1, 0, 0, false)
	p.Grid.AddItem(p.statusBar, 3, 0, 1, 1, 0, 0, false)

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
	p.table.SetCell(0, 2, tview.NewTableCell("Score").SetTextColor(tcell.ColorYellow))
	p.table.SetCell(0, 3, tview.NewTableCell("Tags").SetTextColor(tcell.ColorYellow))

	// Add results to table
	for i, subject := range p.results {
		row := i + 1
		// Set the subject ID as the reference for the first cell
		p.table.SetCell(row, 0, tview.NewTableCell(subject.GetName()).SetReference(int(subject.ID)))

		p.table.SetCell(row, 1, tview.NewTableCell(typeToString(int(subject.Type))))

		// Format score
		scoreStr := fmt.Sprintf("%.1f", subject.Rating.Score)
		p.table.SetCell(row, 2, tview.NewTableCell(scoreStr))

		// Join tags
		var tagNames []string
		for _, tag := range subject.Tags {
			tagNames = append(tagNames, tag.Name)
		}
		p.table.SetCell(row, 3, tview.NewTableCell(strings.Join(tagNames[:min(3, len(tagNames))], ", ")))
	}
}

func (p *SearchPage) makeInputCapture(next, prev tview.Primitive, enterTriggersFetch bool) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			if enterTriggersFetch {
				p.fetchData()
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
	p.tagInput.SetInputCapture(p.makeInputCapture(p.table, p.searchInput, true))

	p.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyBacktab {
			p.app.SetFocus(p.tagInput)
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

func (p *SearchPage) GetName() string {
	return "search"
}

// Convert subject type to string
func typeToString(t int) string {
	var typeStr string
	switch t {
	case 1:
		typeStr = "Book"
	case 2:
		typeStr = "Anime"
	case 3:
		typeStr = "Music"
	case 4:
		typeStr = "Game"
	case 6:
		typeStr = "Real"
	default:
		typeStr = "Unknown"
	}
	return typeStr
}
