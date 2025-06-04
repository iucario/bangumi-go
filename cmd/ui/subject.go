package tui

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd/subject"
	"github.com/rivo/tview"
)

type SubjectPage struct {
	*tview.Grid
	client  api.Client
	app     *App
	Subject *api.Subject
	// Optional
	Collection *api.UserSubjectCollection
}

func NewSubjectPage(a *App, ID int) *SubjectPage {
	sbj := subject.GetSubjectInfo(a.User.Client, ID)

	// Get user collection data for this subject
	var collection *api.UserSubjectCollection
	if a.User != nil && a.User.Client != nil && a.User.Username != "" {
		c, err := subject.GetUserSubjectCollection(a.User.Client, a.User.Username, ID)
		if err == nil && c.Type != 0 {
			collection = &c
		} else {
			collection = &api.UserSubjectCollection{
				SubjectType: sbj.Type,
				Subject:     sbj.ToSlimSubject(),
				Type:        uint32(api.CollectionType[api.Watching]),
				SubjectID:   sbj.ID,
			}
		}
	}

	sub := &SubjectPage{
		Grid:       tview.NewGrid(),
		app:        a,
		Subject:    sbj,
		client:     a.User.Client,
		Collection: collection,
	}
	sub.render()
	sub.setKeyBindings()
	return sub
}

// render displays the subject information and user collection data if available.
func (s *SubjectPage) render() {
	s.SetRows(1, 0, 1)
	s.SetColumns(-1)
	s.SetBorder(false)
	top := tview.NewTextView().SetTextAlign(tview.AlignCenter)
	top.SetText(s.Subject.GetName())

	content := s.createContentTable()

	footer := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	footer.SetText("e: 编辑  q: 返回  R: 刷新  ?: Help")

	s.AddItem(top, 0, 0, 1, 1, 0, 0, false).
		AddItem(content, 1, 0, 1, 1, 0, 0, true).
		AddItem(footer, 2, 0, 1, 1, 0, 0, false)
	s.SetBorders(true)
}

func (s *SubjectPage) Refresh() {
	slog.Debug("subject refresh")
	s.Subject = subject.GetSubjectInfo(s.client, int(s.Subject.ID))
	if s.app.User != nil && s.app.User.Client != nil && s.app.User.Username != "" {
		c, err := subject.GetUserSubjectCollection(s.app.User.Client, s.app.User.Username, int(s.Subject.ID))
		if err == nil && c.Type != 0 {
			s.Collection = &c
		} else {
			s.Collection = nil
		}
	}
	// Re-render the page
	s.Clear()
	s.render()
}

func (s *SubjectPage) setKeyBindings() {
	slog.Debug("setKeyBindings called")
	s.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'e':
			modal := NewEditModal(s.app, *s.Collection, s.onSave)
			s.app.Pages.AddPage("collect", modal, true, true)
			s.app.SetFocus(modal)
		case 'q':
			// Remove subject page. Go back to previous page if any
			s.app.GoBack()
		case 'R':
			s.Refresh()
		default:
			if s.app != nil {
				s.app.handlePageSwitch(event.Rune())
			}
		}
		return event
	})
}

func (s *SubjectPage) onSave(collection *api.UserSubjectCollection) {
	slog.Debug("Save Subject", "collect", collection)
	if collection == nil {
		slog.Error("collecting nil subject")
		return
	}
	err := subject.PostCollection(
		s.app.User.Client,
		int(s.Subject.ID),
		collection.GetStatus(),
		collection.Tags,
		collection.Comment,
		int(collection.Rate),
		collection.Private,
	)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to post collection: %v", err))
	}
	subject.WatchToEpisode(s.app.User.Client, int(s.Subject.ID), int(collection.EpStatus))

	// TODO: update collection info
	s.Collection = collection
	// Re-render the page
	s.Clear()
	s.render()
}

func (s *SubjectPage) createContentTable() *tview.Table {
	table := tview.NewTable().SetBorders(false)
	table.SetSeparator(tview.Borders.Vertical)

	// Helper to add a row with right-aligned name and left-aligned value
	addRow := func(row int, name, value string) {
		nameCell := tview.NewTableCell(name).
			SetAlign(tview.AlignRight).
			SetTextColor(tcell.ColorGray)
		valueCell := tview.NewTableCell(value).
			SetAlign(tview.AlignLeft).
			SetTextColor(tcell.ColorWhite)
		table.SetCell(row, 0, nameCell)
		table.SetCell(row, 1, valueCell)
	}

	totalEps := fmt.Sprintf("%d", s.Subject.Eps)
	if s.Subject.Eps == 0 {
		totalEps = "未知"
	}

	row := 0
	addRow(row, "中文名", s.Subject.NameCn)
	row++
	addRow(row, "原名", s.Subject.Name)
	row++
	addRow(row, "链接", fmt.Sprintf("https://bgm.tv/subject/%d", s.Subject.ID))
	row++
	addRow(row, "类型", api.SubjectTypeRev[int(s.Subject.Type)])
	row++
	if s.Subject.Nsfw {
		addRow(row, "NSFW", "是")
		row++
	}
	addRow(row, "集数", totalEps)
	row++
	if s.Subject.Volumes > 0 {
		addRow(row, "卷数", fmt.Sprintf("%d", s.Subject.Volumes))
		row++
	}
	addRow(row, "评分", fmt.Sprintf("%.1f", s.Subject.Rating.Score))
	row++
	addRow(row, "排名", fmt.Sprintf("%d", s.Subject.Rating.Rank))
	row++
	addRow(row, "评分人数", fmt.Sprintf("%d", s.Subject.Rating.Total))
	row++
	addRow(row, "标签", renderTags(s.Subject.Tags, s.Subject.WikiTags))
	row++

	// Section: 收藏人数
	labelCell := tview.NewTableCell("[yellow]收藏人数[-]").SetAlign(tview.AlignCenter).SetSelectable(false)
	table.SetCell(row, 0, labelCell)
	table.SetCell(row, 1, tview.NewTableCell("").SetSelectable(false))
	row++
	addRow(row, "在看", fmt.Sprintf("%d", s.Subject.CollectionCount.Watching))
	row++
	addRow(row, "想看", fmt.Sprintf("%d", s.Subject.CollectionCount.Wish))
	row++
	addRow(row, "看过", fmt.Sprintf("%d", s.Subject.CollectionCount.Done))
	row++
	addRow(row, "搁置", fmt.Sprintf("%d", s.Subject.CollectionCount.OnHold))
	row++
	addRow(row, "抛弃", fmt.Sprintf("%d", s.Subject.CollectionCount.Dropped))
	row++

	addRow(row, "放送日期", s.Subject.Date)
	row++
	addRow(row, "简介", s.Subject.Summary)
	row++

	// Section: 你的收藏信息
	if s.Collection != nil && s.Collection.Type != 0 {
		table.SetCell(row, 0, tview.NewTableCell("[yellow]你的收藏信息[-]").SetAlign(tview.AlignCenter).SetSelectable(false))
		table.SetCell(row, 1, tview.NewTableCell("").SetSelectable(false))
		row++
		addRow(row, "状态", string(api.CollectionTypeRev[int(s.Collection.Type)]))
		row++
		addRow(row, "评分", fmt.Sprintf("%d", s.Collection.Rate))
		row++
		addRow(row, "短评", s.Collection.Comment)
		row++
		addRow(row, "标签", strings.Join(s.Collection.Tags, ", "))
		row++
		addRow(row, "看到第", fmt.Sprintf("%d/%d 集", s.Collection.EpStatus, s.Subject.Eps))
		row++
		if s.Collection.VolStatus > 0 {
			addRow(row, "看到第", fmt.Sprintf("%d 卷", s.Collection.VolStatus))
			row++
		}
		addRow(row, "隐私收藏", fmt.Sprintf("%v", s.Collection.Private))
		row++
	}

	return table
}

// renderTags formats the subject tags for display, highlighting wiki tags.
func renderTags(tags []api.Tag, wikiTags []string) string {
	arr := make([]string, 0, len(tags))
	wiki := make(map[string]bool)
	for _, tag := range wikiTags {
		wiki[tag] = true
	}
	for _, tag := range tags {
		text := fmt.Sprintf("%s•%d", tag.Name, tag.Count)
		if _, ok := wiki[tag.Name]; ok {
			text = fmt.Sprintf("[blue]%s[-]", text)
		}
		arr = append(arr, text)
	}
	return "| " + strings.Join(arr, " | ") + " |"
}
