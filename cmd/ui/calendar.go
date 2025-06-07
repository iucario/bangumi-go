package tui

import (
	"log/slog"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd/calendar"
	"github.com/iucario/bangumi-go/internal/ui"
	"github.com/rivo/tview"
)

type CalendarPage struct {
	*tview.Grid
	client *api.HTTPClient
	app    *App
	data   []api.Calendar
	table  *tview.Table
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
	c.SetRows(1, -1, 1)
	c.SetColumns(-1) // One full-width column
	header := tview.NewTextView().
		SetText("放送日历").
		SetTextAlign(tview.AlignCenter).
		SetTextColor(ui.Styles.TitleColor)
	table := tview.NewTable().SetSelectable(true, true)
	c.table = table
	table.SetBorder(false)
	table.SetFixed(1, 0)

	// Map weekday ID to column (0=Monday, 6=Sunday)
	weekdayNames := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	today := int(time.Now().Weekday())
	if today == 0 {
		today = 6 // Sunday
	} else {
		today-- // Monday=0
	}

	// Rotate weekdayNames so that today is first
	rotatedWeekdayNames := append(weekdayNames[today:], weekdayNames[:today]...)

	// Build mapping from weekday ID to column, with today as first column
	weekdayToCol := make(map[uint32]int)
	for i := range 7 {
		// Weekday ID: 1=Mon, ..., 7=Sun
		id := uint32((today+i)%7 + 1)
		weekdayToCol[id] = i
	}

	// Header row: weekday names (today is first column)
	for col, name := range rotatedWeekdayNames {
		color := ui.Styles.PrimaryTextColor
		bgcolor := ui.Styles.ContrastBackgroundColor
		attr := tcell.AttrNone
		if col == 0 {
			attr = tcell.AttrBold
			bgcolor = ui.Styles.MoreContrastBackgroundColor
		}
		table.SetCell(0, col, tview.NewTableCell(name).
			SetTextColor(color).
			SetBackgroundColor(bgcolor).
			SetSelectable(true).
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
				// Limit name width
				maxWidth := 10
				runeName := []rune(name)
				if len(runeName) > maxWidth {
					name = string(runeName[:maxWidth-1]) + "…"
				}
				cell := tview.NewTableCell(name).
					SetReference(anime.ID).
					SetTextColor(ui.Styles.PrimaryTextColor).
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

	footer := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	footer.SetText("Enter: 详情  ←/→ ↑/↓: 移动  ?: Help")

	c.Clear()
	c.AddItem(header, 0, 0, 1, 1, 0, 0, false).
		AddItem(table, 1, 0, 1, 1, 0, 0, true).
		AddItem(footer, 2, 0, 1, 1, 0, 0, false)
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
