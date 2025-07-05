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
	client      *api.HTTPClient
	app         *App
	table       *tview.Table
	searchInput *tview.InputField
	tagInput    *tview.InputField
	// Date picker fields (start and end)
	startYearInput *tview.InputField
	endYearInput   *tview.InputField
	results        []api.Subject
	typeCheckboxes []*tview.Checkbox
	statusBar      *ui.StatusBar
	// Pagination fields
	currentPage  int
	pageSize     int
	totalResults int
}

func NewSearchPage(app *App) *SearchPage {
	typeLabels := api.S_TYPE_ALL
	typeCheckboxes := make([]*tview.Checkbox, len(typeLabels))
	for i, label := range typeLabels {
		cb := tview.NewCheckbox().SetLabel(label.CN())
		typeCheckboxes[i] = cb
	}
	// Date input fields
	startDate := tview.NewInputField().SetLabel("日期: ").SetFieldWidth(10).SetPlaceholder("YYYYMMDD")
	endDate := tview.NewInputField().SetLabel("- ").SetFieldWidth(10).SetPlaceholder("YYYYMMDD")
	// Only allow numbers
	numberOnly := func(textToCheck string, lastChar rune) bool {
		if lastChar == 0 {
			return true
		}
		return lastChar >= '0' && lastChar <= '9'
	}
	startDate.SetAcceptanceFunc(numberOnly)
	endDate.SetAcceptanceFunc(numberOnly)
	search := &SearchPage{
		Grid:           tview.NewGrid(),
		client:         app.User.Client.HTTPClient,
		app:            app,
		table:          tview.NewTable(),
		searchInput:    tview.NewInputField().SetLabel("关键词: ").SetFieldWidth(40),
		tagInput:       tview.NewInputField().SetLabel("标签  : ").SetFieldWidth(40),
		startYearInput: startDate,
		endYearInput:   endDate,
		typeCheckboxes: typeCheckboxes,
		statusBar:      ui.NewStatusBar(),
		currentPage:    1,
		pageSize:       20,
	}
	search.render()
	search.setKeyBindings()
	app.SetFocus(search.searchInput)
	return search
}

// search fetches data and render the ui accordingly
func (p *SearchPage) search() {
	p.statusBar.SetMessage("Searching...", "info")
	p.currentPage = 1 // Always reset to first page on new search
	p.fetchData()
	maxPage := 1
	if p.pageSize > 0 {
		maxPage = (p.totalResults + p.pageSize - 1) / p.pageSize
	}
	statusMsg := fmt.Sprintf("Found %d results (Page %d/%d)", p.totalResults, p.currentPage, maxPage)
	p.statusBar.SetMessage(statusMsg, "success")
	p.render()
}

// paginateSearch fetches data and renders the UI for the current page (without resetting page number)
func (p *SearchPage) paginateSearch() {
	p.statusBar.SetMessage("Searching...", "info")
	p.fetchData()
	maxPage := 1
	if p.pageSize > 0 {
		maxPage = (p.totalResults + p.pageSize - 1) / p.pageSize
	}
	statusMsg := fmt.Sprintf("Found %d results (Page %d/%d)", p.totalResults, p.currentPage, maxPage)
	p.statusBar.SetMessage(statusMsg, "success")
	p.render()
}

func (p *SearchPage) fetchData() {
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
	// Date range
	if len(p.startYearInput.GetText()) == 8 {
		ymd := p.startYearInput.GetText()
		startDate := fmt.Sprintf(">=%s-%s-%s", ymd[:4], ymd[4:6], ymd[6:8])
		filter.AirDate = append(filter.AirDate, startDate)
	}
	if len(p.endYearInput.GetText()) == 8 {
		ymd := p.endYearInput.GetText()
		endDate := fmt.Sprintf("<=%s-%s-%s", ymd[:4], ymd[4:6], ymd[6:8])
		filter.AirDate = append(filter.AirDate, endDate)
	}
	if keyword == "" && tags == "" && len(selectedTypes) == 0 && len(filter.AirDate) == 0 {
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
	pagesize := p.pageSize
	offset := (p.currentPage - 1) * p.pageSize
	result, err := search.Search(p.client, payload, pagesize, offset)
	if err != nil {
		p.statusBar.SetMessage(fmt.Sprintf("Error searching: %v error", err.Error()), "error")
		return
	}
	p.results = result.Data
	p.totalResults = result.Total
	if p.totalResults == 0 {
		p.totalResults = len(result.Data)
	}
}

func (p *SearchPage) render() {
	p.searchInput.SetLabel("关键词: ").SetFieldWidth(40)
	p.tagInput.SetLabel("标签:   ").SetFieldWidth(40)
	// Use flexible row heights for better layout
	p.Grid.SetRows(2, 2, 2, 2, 0, 1)
	p.Grid.SetColumns(0)
	dateFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	dateFlex.AddItem(p.startYearInput, 18, 0, false)
	dateFlex.AddItem(p.endYearInput, 24, 0, false)

	// Create typeFlex locally
	typeFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	for _, cb := range p.typeCheckboxes {
		typeFlex.AddItem(cb, 8, 0, false)
	}
	// Add items to grid with correct row indices
	p.Grid.Clear()
	p.Grid.AddItem(p.searchInput, 0, 0, 1, 1, 0, 0, true)
	p.Grid.AddItem(p.tagInput, 1, 0, 1, 1, 0, 0, false)
	p.Grid.AddItem(dateFlex, 2, 0, 1, 1, 0, 0, false)
	p.Grid.AddItem(typeFlex, 3, 0, 1, 1, 0, 0, false)
	p.Grid.AddItem(p.table, 4, 0, 1, 1, 0, 0, false)
	p.Grid.AddItem(p.statusBar, 5, 0, 1, 1, 0, 0, false)

	// Set up the results table
	p.table.Clear() // Clear all previous rows to avoid leftover data
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

// makeInputCapture defines the previous and next primitive to focus
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
	// Tab order: searchInput -> tagInput -> startYear -> startMonth -> startDay -> endYear -> endMonth -> endDay -> typeCheckboxes[0] ...
	p.searchInput.SetInputCapture(p.makeInputCapture(p.tagInput, nil, true))
	p.tagInput.SetInputCapture(p.makeInputCapture(p.startYearInput, p.searchInput, true))
	p.startYearInput.SetInputCapture(p.makeInputCapture(p.endYearInput, p.tagInput, true))
	p.endYearInput.SetInputCapture(p.makeInputCapture(p.typeCheckboxes[0], p.startYearInput, true))
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
					p.app.SetFocus(p.endYearInput)
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
						p.app.OpenSubjectPage(subjectID, "search")
						return nil
					}
				}
			}
			return nil
		}
		if event.Key() == tcell.KeyRune {
			// Pagination: n for next, p for previous
			switch event.Rune() {
			case 'n':
				maxPage := (p.totalResults + p.pageSize - 1) / p.pageSize
				if p.currentPage < maxPage {
					p.currentPage++
					p.paginateSearch()
				}
				return nil
			case 'p':
				if p.currentPage > 1 {
					p.currentPage--
					p.paginateSearch()
				}
				return nil
			default:
				p.app.handlePageSwitch(event.Rune())
			}
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

func (p *SearchPage) OnFocus() {
	p.app.SetFocus(p.table)
}

// Focus implements tview.Primitive's Focus method for correct page focus handling.
func (p *SearchPage) Focus(delegate func(tview.Primitive)) {
	// Custom focus logic: select table if results, else search input
	if len(p.results) > 0 {
		// Ensure a row is selected
		if p.table.GetRowCount() > 1 {
			row, _ := p.table.GetSelection()
			if row <= 0 {
				p.table.Select(1, 0)
			}
		}
		delegate(p.table)
	} else {
		delegate(p.searchInput)
	}
}
