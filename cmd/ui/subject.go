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
	content *tview.TextView
	// Optional
	Collection *api.UserSubjectCollection
}

func NewSubjectPage(a *App, ID int) *SubjectPage {
	data := subject.GetSubjectInfo(a.User.Client, ID)

	// Get user collection data for this subject
	var collection *api.UserSubjectCollection
	if a.User != nil && a.User.Client != nil && a.User.Username != "" {
		c, err := subject.GetUserSubjectCollection(a.User.Client, a.User.Username, ID)
		if err == nil && c.Type != 0 {
			collection = &c
		} else {
			collection = nil
		}
	}

	sub := &SubjectPage{
		Grid:       tview.NewGrid(),
		app:        a,
		Subject:    data,
		client:     a.User.Client,
		Collection: collection,
	}
	sub.render()
	sub.setKeyBindings()
	return sub
}

// render displays the subject information and user collection data if available.
func (s *SubjectPage) render() {
	s.SetRows(2, 0, 1)
	s.SetColumns(-1)
	s.SetBorder(false)
	top := tview.NewTextView().SetTextAlign(tview.AlignCenter)
	top.SetText(s.Subject.GetName())

	text := s.createText()
	s.content = tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetWrap(true)
	s.content.SetText(text)
	s.content.SetBorder(true)
	s.content.SetBorderColor(tcell.ColorYellow)

	footer := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	footer.SetText("e: 编辑  q: 返回  R: 刷新  ?: Help")

	s.AddItem(top, 0, 0, 1, 1, 0, 0, false).
		AddItem(s.content, 1, 0, 1, 1, 0, 0, true).
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
	s.content.SetText(s.createText())
}

func (s *SubjectPage) setKeyBindings() {
	slog.Debug("setKeyBindings called")
	s.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'e':
			modal := NewCollectModal(s.app, s.Subject, s.onSave)
			s.app.Pages.AddPage("collect", modal, true, true)
			s.app.SetFocus(modal)
		case 'q':
			// Remove subject page. Go back to previous page if any
			if prev, ok := s.app.PopPage(); ok {
				s.app.Pages.SwitchToPage(prev)
			}
			s.app.Pages.RemovePage("subject")
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
		slog.Error("collecting subject")
		return
	}
	// TODO: update collection info
	s.Collection = collection
	s.content.SetText(s.createText())
}

func (s *SubjectPage) createText() string {
	// Compose subject info
	totalEps := fmt.Sprintf("%d", s.Subject.Eps)
	if s.Subject.Eps == 0 {
		totalEps = "未知"
	}
	text := fmt.Sprintf("[yellow]%s[-]\n%s\n", s.Subject.NameCn, s.Subject.Name)
	text += fmt.Sprintf("https://bgm.tv/subject/%d\n", s.Subject.ID)
	text += "\n--------\n"
	text += fmt.Sprintf("[yellow]类型[-]: %s\n", api.SubjectTypeRev[int(s.Subject.Type)])
	if s.Subject.Nsfw {
		text += "[red]NSFW[-]\n"
	}
	text += fmt.Sprintf("[yellow]集数[-]: %s\n", totalEps)
	if s.Subject.Volumes > 0 {
		text += fmt.Sprintf("[yellow]卷数:[-] %d\n", s.Subject.Volumes)
	}
	text += fmt.Sprintf("[yellow]评分[-]: %.1f\n", s.Subject.Rating.Score)
	text += fmt.Sprintf("[yellow]排名[-]: %d\n", s.Subject.Rating.Rank)
	text += fmt.Sprintf("[yellow]评分人数[-]: %d\n", s.Subject.Rating.Total)
	text += fmt.Sprintf("[yellow]标签[-]: %s\n", renderTags(s.Subject.Tags, s.Subject.WikiTags))
	text += "\n收藏人数\n"
	text += fmt.Sprintf("[yellow]在看[-]: %d\n", s.Subject.CollectionCount.Watching)
	text += fmt.Sprintf("[yellow]想看[-]: %d\n", s.Subject.CollectionCount.Wish)
	text += fmt.Sprintf("[yellow]看过[-]: %d\n", s.Subject.CollectionCount.Done)
	text += fmt.Sprintf("[yellow]搁置[-]: %d\n", s.Subject.CollectionCount.OnHold)
	text += fmt.Sprintf("[yellow]抛弃[-]: %d\n", s.Subject.CollectionCount.Dropped)
	text += "\n--------\n"
	text += fmt.Sprintf("[yellow]放送日期[-]: %s\n", s.Subject.Date)
	text += fmt.Sprintf("[yellow]简介[-]: \n%s\n", s.Subject.Summary)

	// Show user collection info if available
	if s.Collection != nil && s.Collection.Type != 0 {
		text += "\n[yellow]你的收藏信息[-]:\n"
		text += fmt.Sprintf("状态: %s\n", api.CollectionTypeRev[int(s.Collection.Type)])
		text += fmt.Sprintf("评分: %d\n", s.Collection.Rate)
		text += fmt.Sprintf("短评: %s\n", s.Collection.Comment)
		text += fmt.Sprintf("标签: %s\n", strings.Join(s.Collection.Tags, ", "))
		text += fmt.Sprintf("看到第 %d 集\n", s.Collection.EpStatus)
		if s.Collection.VolStatus > 0 {
			text += fmt.Sprintf("看到第 %d 卷\n", s.Collection.VolStatus)
		}
	}
	return text
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
