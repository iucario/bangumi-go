package tui

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd/subject"
	"github.com/iucario/bangumi-go/internal/ui"
	"github.com/rivo/tview"
)

type SubjectPage struct {
	*tview.Grid
	client       api.Client
	app          *App
	Subject      *api.Subject
	leftContent  *tview.TextView
	rightContent *tview.TextView
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
				Type:        0,
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
// There is a header, footer and content in the middle.
// Content has two parts, left and right content
func (s *SubjectPage) render() {
	s.SetRows(1, 0, 1)
	s.SetColumns(40, -1)
	s.SetBorder(false)
	s.SetBorders(false)
	s.SetBorderColor(tcell.ColorGray)
	top := tview.NewTextView().SetTextAlign(tview.AlignCenter)
	top.SetText(fmt.Sprintf("%s %s", s.Subject.GetName(), api.SubjectTypeRev[int(s.Subject.Type)]))

	text := s.createLeftText()
	rightText := s.createRightText()
	s.leftContent = tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetWrap(true)
	s.rightContent = tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetWrap(true)
	s.leftContent.SetText(text)
	s.rightContent.SetText(rightText)
	// Initially, leftContent is focused, so show its border
	s.leftContent.SetBorder(true)
	s.rightContent.SetBorder(true)

	// Change border color on focus/blur
	s.leftContent.SetFocusFunc(func() {
		s.leftContent.SetBorderColor(ui.Styles.TitleColor) // Focused color
	})
	s.leftContent.SetBlurFunc(func() {
		s.leftContent.SetBorderColor(tcell.ColorGray) // Unfocused color
	})
	s.rightContent.SetFocusFunc(func() {
		s.rightContent.SetBorderColor(ui.Styles.TitleColor)
	})
	s.rightContent.SetBlurFunc(func() {
		s.rightContent.SetBorderColor(tcell.ColorGray)
	})
	footer := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	footer.SetText("e: 编辑  q: 返回  R: 刷新 ←/→: 移动 ↑/↓: 滚动 ?: Help")

	s.AddItem(top, 0, 0, 1, 2, 0, 0, false).
		AddItem(s.leftContent, 1, 0, 1, 1, 0, 0, false).
		AddItem(s.rightContent, 1, 1, 1, 1, 0, 0, true).
		AddItem(footer, 2, 0, 1, 2, 0, 0, false)
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
	s.leftContent.SetText(s.createLeftText())
	s.rightContent.SetText(s.createRightText())
}

func (s *SubjectPage) setKeyBindings() {
	slog.Debug("setKeyBindings called")
	s.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
			s.app.SetFocus(s.leftContent)
		case tcell.KeyRight:
			s.app.SetFocus(s.rightContent)
		default:
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
	s.leftContent.SetText(s.createLeftText())
	s.rightContent.SetText(s.createRightText())
}

func (s *SubjectPage) createLeftText() string {
	// Compose subject info
	totalEps := fmt.Sprintf("%d", s.Subject.Eps)
	if s.Subject.Eps == 0 {
		totalEps = "未知"
	}
	text := fmt.Sprintf("%s\n%s\n", ui.SecondaryText(s.Subject.NameCn), s.Subject.Name)
	text += fmt.Sprintf("https://bgm.tv/subject/%d\n", s.Subject.ID)
	if s.Subject.Nsfw {
		text += ui.GraphicsColor("NSFW") + "\n"
	}
	text += fmt.Sprintf("集数: %s\n", totalEps)
	if s.Subject.Volumes > 0 {
		text += fmt.Sprintf("卷数: %d\n", s.Subject.Volumes)
	}
	text += fmt.Sprintf("评分: %.1f\n", s.Subject.Rating.Score)
	text += fmt.Sprintf("排名: %d\n", s.Subject.Rating.Rank)
	text += fmt.Sprintf("评分人数: %d\n", s.Subject.Rating.Total)
	text += ui.SecondaryText("\n收藏人数\n")
	text += fmt.Sprintf("在看: %d\n", s.Subject.CollectionCount.Watching)
	text += fmt.Sprintf("想看: %d\n", s.Subject.CollectionCount.Wish)
	text += fmt.Sprintf("看过: %d\n", s.Subject.CollectionCount.Done)
	text += fmt.Sprintf("搁置: %d\n", s.Subject.CollectionCount.OnHold)
	text += fmt.Sprintf("抛弃: %d\n", s.Subject.CollectionCount.Dropped)
	text += "\n"
	text += fmt.Sprintf("放送日期: %s\n", ui.SecondaryText(s.Subject.Date))

	// Show user collection info if available
	if s.Collection != nil && s.Collection.Type != 0 {
		text += ui.SecondaryText("\n你的收藏信息:\n")
		text += fmt.Sprintf("状态: %s\n", s.Collection.GetStatus())
		text += fmt.Sprintf("评分: %d\n", s.Collection.Rate)
		text += fmt.Sprintf("短评: %s\n", s.Collection.Comment)
		text += fmt.Sprintf("标签: %s\n", strings.Join(s.Collection.Tags, ", "))
		text += fmt.Sprintf("看到第 %d 集\n", s.Collection.EpStatus)
		if s.Collection.VolStatus > 0 {
			text += fmt.Sprintf("看到第 %d 卷\n", s.Collection.VolStatus)
		}
		text += fmt.Sprintf("隐私收藏: %v\n", s.Collection.Private)
	}
	return text
}

func (s *SubjectPage) createRightText() string {
	text := ""
	text += fmt.Sprintf("%s\n", s.Subject.Summary)
	text += "\n\n"
	text += fmt.Sprintf("标签: %s\n", renderTags(s.Subject.Tags, s.Subject.WikiTags))
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
			text = fmt.Sprint(ui.TitleColor(text))
		}
		arr = append(arr, text)
	}
	return "| " + strings.Join(arr, " | ") + " |"
}
