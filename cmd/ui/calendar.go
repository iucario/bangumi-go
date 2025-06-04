package tui

import (
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd/calendar"
	"github.com/rivo/tview"
)

type CalendarPage struct {
	*tview.Grid
	client *api.HTTPClient
	data   []api.Calendar
	app    *App
}

func NewCalendarPage(app *App) *CalendarPage {
	calendar := &CalendarPage{
		Grid:   tview.NewGrid(),
		client: api.NewHTTPClient(""),
		app:    app,
	}
	calendar.fetchData()
	calendar.render()
	calendar.setKeyBindings()
	return calendar
}

func (c *CalendarPage) fetchData() {
	calendars, err := calendar.GetCalendar(c.client)
	if err != nil {
		slog.Error("Failed to fetch calendar data", "error", err)
		return
	}
	c.data = calendars
}

func (c *CalendarPage) render() {
	c.SetRows(-1)    // Divide all available space equally
	c.SetColumns(-1) // One full-width column
	table := tview.NewTable().SetSelectable(true, true)
	table.SetBorder(true).SetTitle("Anime Calendar")

	// Map weekday ID to column (0=Monday, 6=Sunday)
	weekdayToCol := map[uint32]int{1: 0, 2: 1, 3: 2, 4: 3, 5: 4, 6: 5, 7: 6}
	colToWeekday := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

	// Find today's weekday column (Go: Sunday=0, our mapping: Monday=0)
	today := int(time.Now().Weekday())
	if today == 0 {
		today = 6 // Sunday
	} else {
		today-- // Monday=0
	}

	// Header row: weekday names
	for col, name := range colToWeekday {
		color := tcell.ColorWhite
		attr := tcell.AttrBold
		if col == today {
			color = tcell.ColorYellow
			attr = tcell.AttrBold | tcell.AttrReverse
		}
		table.SetCell(0, col, tview.NewTableCell(name).
			SetTextColor(color).
			SetSelectable(false).
			SetAlign(tview.AlignCenter).
			SetAttributes(attr))
	}

	// Prepare anime lists for each weekday column
	maxRows := 0
	animeByCol := make([][]api.CalendarItem, 7)
	for _, cal := range c.data {
		col, ok := weekdayToCol[cal.Weekday.ID]
		if !ok || col < 0 || col > 6 {
			continue
		}
		// Sort by followers
		sort.Slice(cal.Items, func(i, j int) bool {
			fi := cal.Items[i].CollectionCount.Wish + cal.Items[i].CollectionCount.Watching + cal.Items[i].CollectionCount.Done
			fj := cal.Items[j].CollectionCount.Wish + cal.Items[j].CollectionCount.Watching + cal.Items[j].CollectionCount.Done
			return fi > fj
		})
		animeByCol[col] = cal.Items
		if len(cal.Items) > maxRows {
			maxRows = len(cal.Items)
		}
	}

	// Fill table with anime rows (start from row 1)
	for row := range maxRows {
		for col := range 7 {
			if row < len(animeByCol[col]) {
				anime := animeByCol[col][row]
				name := anime.Name
				if anime.NameCn != "" {
					name = anime.NameCn
				}
				followers := anime.CollectionCount.Wish + anime.CollectionCount.Watching + anime.CollectionCount.Done
				cellText := fmt.Sprintf("%d+%s", followers, name)
				cell := tview.NewTableCell(cellText).
					SetReference(anime.ID).
					SetTextColor(tcell.ColorGreen).
					SetAlign(tview.AlignLeft)
				table.SetCell(row+1, col, cell)
			} else {
				table.SetCell(row+1, col, tview.NewTableCell("").SetSelectable(false))
			}
		}
	}

	table.SetSelectedFunc(func(row, column int) {
		cell := table.GetCell(row, column)
		if cell == nil || cell.GetReference() == nil {
			return
		}
		if c.app != nil {
			if subjectID, ok := cell.GetReference().(int); ok {
				c.app.OpenSubjectPage(subjectID, "calendar")
			}
		}
	})

	c.Clear()
	c.AddItem(table, 0, 0, 1, 1, 0, 0, true)
}

func (c *CalendarPage) setKeyBindings() {
	c.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		default:
			if c.app != nil {
				c.app.handlePageSwitch(event.Rune())
			}
		}
		return event
	})
}
