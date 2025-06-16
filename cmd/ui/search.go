package tui

import (
	"github.com/iucario/bangumi-go/api"
	"github.com/rivo/tview"
)

type SearchPage struct {
	*tview.Grid
	client *api.HTTPClient
	app    *App
	table  *tview.Table
}

func NewSearchPage(app *App) *SearchPage {
	search := &SearchPage{
		Grid:   tview.NewGrid(),
		client: api.NewHTTPClient(""),
		app:    app,
	}
	// search.fetchData()
	// search.render()
	// search.setKeyBindings()
	return search
}
