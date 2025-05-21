package tui

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd/calendar"
	"github.com/rivo/tview"
)

type Calendar struct {
	*tview.Grid
	client *api.HTTPClient
	data   []api.Calendar
	app    *App
}

func NewCalendar(app *App) *Calendar {
	calendar := &Calendar{
		Grid:   tview.NewGrid(),
		client: api.NewHTTPClient(""),
		app:    app,
	}
	calendar.initLayout()
	return calendar
}

func NewCalendarPage(app *App) *Calendar {
	calendar := NewCalendar(app)
	calendar.fetchData()
	calendar.renderCalendar()
	return calendar
}

func (c *Calendar) initLayout() {
	c.SetRows(-1)     // Divide all available space equally
	c.SetColumns(-1)  // One full-width column
	c.SetBorder(true) // Add a border around the entire calendar
	c.SetTitle("Anime Calendar")
}

func (c *Calendar) fetchData() {
	calendars, err := calendar.GetCalendar(c.client)
	if err != nil {
		slog.Error("Failed to fetch calendar data", "error", err)
		return
	}
	c.data = calendars
}

func (c *Calendar) renderCalendar() {
	calendarView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)

	weekday := time.Now().Weekday()

	var content strings.Builder
	for _, cal := range c.data {
		// Sort items by follower count
		sort.Slice(cal.Items, func(i, j int) bool {
			followersI := cal.Items[i].CollectionCount.Wish + cal.Items[i].CollectionCount.Watching + cal.Items[i].CollectionCount.Done
			followersJ := cal.Items[j].CollectionCount.Wish + cal.Items[j].CollectionCount.Watching + cal.Items[j].CollectionCount.Done
			return followersI > followersJ
		})

		// Print header with a different color
		weekdayTitle := fmt.Sprintf("\t%d %s", cal.Weekday.ID, cal.Weekday.EN)
		if cal.Weekday.ID == uint32(weekday) {
			content.WriteString(fmt.Sprintf("[yellow]-> %s[white]\n", weekdayTitle))
		} else {
			content.WriteString(fmt.Sprintf("%s\n", weekdayTitle))
		}
		content.WriteString(strings.Repeat("─", 50) + "\n") // Separator line

		// Print items
		for _, anime := range cal.Items {
			name := anime.Name
			if anime.NameCn != "" {
				name = anime.NameCn
			}
			followers := anime.CollectionCount.Wish + anime.CollectionCount.Watching + anime.CollectionCount.Done
			content.WriteString(fmt.Sprintf("[green]%6d[white] │ %s\n", followers, name))
		}
		content.WriteString("\n")
	}

	calendarView.SetText(content.String())

	// Add scrolling support with enhanced navigation
	calendarView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := calendarView.GetScrollOffset()
		_, _, _, height := calendarView.GetInnerRect()
		pageSize := height - 1

		switch event.Key() {
		case tcell.KeyUp, tcell.KeyCtrlP:
			calendarView.ScrollTo(row-1, 0)
		case tcell.KeyDown, tcell.KeyCtrlN:
			calendarView.ScrollTo(row+1, 0)
		case tcell.KeyPgUp:
			calendarView.ScrollTo(row-pageSize, 0)
		case tcell.KeyPgDn:
			calendarView.ScrollTo(row+pageSize, 0)
		case tcell.KeyHome:
			calendarView.ScrollToBeginning()
		case tcell.KeyEnd:
			calendarView.ScrollToEnd()
		}

		// Support vim-style navigation
		switch event.Rune() {
		case 'k':
			calendarView.ScrollTo(row-1, 0)
		case 'j':
			calendarView.ScrollTo(row+1, 0)
		case 'g':
			calendarView.ScrollToBeginning()
		case 'G':
			calendarView.ScrollToEnd()
		case 'b', 'u':
			calendarView.ScrollTo(row-pageSize, 0)
		case 'f', 'd':
			calendarView.ScrollTo(row+pageSize, 0)
		default:
			if c.app != nil {
				c.app.handlePageSwitch(event.Rune())
			}
		}
		return event
	})

	c.AddItem(calendarView, 0, 0, 1, 1, 0, 0, true)
}
