package tui

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd/subject"
	"github.com/iucario/bangumi-go/internal/task"
	"github.com/iucario/bangumi-go/internal/ui"
	"github.com/rivo/tview"
)

type SubjectPage struct {
	*tview.Grid
	client       api.Client
	app          *App
	Subject      *api.Subject
	Episodes     *api.Episodes
	leftContent  *tview.TextView
	rightContent *tview.TextView
	// Optional
	Collection *api.UserSubjectCollection
}

func NewSubjectPage(a *App, ID int) *SubjectPage {
	// Get data concurrently
	tasks := []task.Task{
		{
			ID: "subject",
			Do: func() (any, error) {
				return subject.GetSubjectInfo(a.User.Client, ID)
			},
		},
		{
			ID: "collection",
			Do: func() (any, error) {
				return subject.GetUserSubjectCollection(a.User.Client, a.User.Username, ID)
			},
		},
		{
			ID: "episodes",
			Do: func() (any, error) {
				return subject.GetEpisodes(a.User.Client.HTTPClient, ID, 0, 100)
			},
		},
	}
	res := task.Run(tasks)

	// TODO: fix this ugly block
	sbjRes, ok := res["subject"]
	if !ok || sbjRes.Error != nil || sbjRes.Data == nil {
		slog.Error("Failed to fetch subject info", "ID", ID, "Error", sbjRes.Error)
		return nil
	}
	sbj := sbjRes.Data.(*api.Subject)

	episodesRes, ok := res["episodes"]
	if !ok || episodesRes.Error != nil || episodesRes.Data == nil {
		slog.Error("Failed to fetch episodes", "Error", episodesRes.Error)
	}
	episodes := episodesRes.Data.(*api.Episodes)

	// Get user collection data for this subject
	collection := &api.UserSubjectCollection{
		SubjectType: sbj.Type,
		Subject:     sbj.SlimSubject,
		Type:        0,
		SubjectID:   sbj.ID,
	}
	if a.User != nil && a.User.Client != nil && a.User.Username != "" {
		c, ok := res["collection"].Data.(api.UserSubjectCollection)
		if ok && c.Type != 0 {
			collection = &c
		} else {
			slog.Error("Failed to fetch collection", "Error", res["collection"].Error)
		}
	}

	sub := &SubjectPage{
		Grid:       tview.NewGrid(),
		app:        a,
		Subject:    sbj,
		Episodes:   episodes,
		client:     a.User.Client,
		Collection: collection,
	}
	sub.render()
	sub.setKeyBindings()
	return sub
}

func (s *SubjectPage) GetName() string {
	return "subject"
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
	top.SetText(fmt.Sprintf("%s %s %s", s.Subject.GetName(), s.Subject.Platform, api.SubjectTypeRev[int(s.Subject.Type)]))
	top.SetTextColor(ui.Styles.TitleColor)

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
	footer.SetText("e: 编辑  q: 返回  R: 刷新  ←/→: 移动  ↑/↓: 滚动  ?: Help")

	s.AddItem(top, 0, 0, 1, 2, 0, 0, false).
		AddItem(s.leftContent, 1, 0, 1, 1, 0, 0, false).
		AddItem(s.rightContent, 1, 1, 1, 1, 0, 0, true).
		AddItem(footer, 2, 0, 1, 2, 0, 0, false)
}

func (s *SubjectPage) Refresh() {
	slog.Debug("subject refresh")
	s.Subject, _ = subject.GetSubjectInfo(s.client, int(s.Subject.ID))
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
	s.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft:
			s.app.SetFocus(s.leftContent)
		case tcell.KeyRight:
			s.app.SetFocus(s.rightContent)
		case tcell.KeyRune:
			switch event.Rune() {
			case 'e':
				modal := NewCollectModal(s.app, *s.Collection, s.onSave)
				s.app.Pages.AddPage("collect", modal, true, true)
				s.app.SetFocus(modal)
			case 'R':
				s.Refresh()
			default:
				s.app.handlePageSwitch(event.Rune())
			}
		}
		return event
	})
}

func (s *SubjectPage) onSave(collection *api.UserSubjectCollection) error {
	slog.Debug("Save Subject", "collect", collection)
	if collection == nil {
		slog.Error("collecting nil subject")
		return errors.New("subject is nil")
	}
	// Find the original collection for comparison
	var original *api.UserSubjectCollection
	if s.Collection != nil && s.Collection.SubjectID == collection.SubjectID {
		original = s.Collection
	} else {
		original = &api.UserSubjectCollection{}
	}

	if CollectionInfoChanged(original, collection) {
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
			return err
		}
	}

	if EpisodeStatusChanged(original, collection) {
		subject.WatchToEpisode(s.app.User.Client, int(s.Subject.ID), int(collection.EpStatus))
	}

	// Update collection info
	s.Collection = collection
	s.leftContent.SetText(s.createLeftText())
	s.rightContent.SetText(s.createRightText())
	return nil
}

func (s *SubjectPage) createLeftText() string {
	// Compose subject info
	totalEps := fmt.Sprintf("%d", s.Subject.Eps)
	if s.Subject.Eps == 0 {
		totalEps = "未知"
	}
	text := ""
	if s.Subject.NameCn != "" && s.Subject.NameCn != s.Subject.Name {
		text += fmt.Sprintf("%s\n", ui.SecondaryText(s.Subject.NameCn))
	}
	text += fmt.Sprintf("%s\n", ui.SecondaryText(s.Subject.Name))
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
		text += fmt.Sprintf("看到第 %d 集/%d\n", s.Collection.EpStatus, s.Subject.Eps)
		text += fmt.Sprintf("评分: %d\n", s.Collection.Rate)
		text += fmt.Sprintf("标签: %s\n", strings.Join(s.Collection.Tags, ", "))
		text += fmt.Sprintf("短评: %s\n", s.Collection.Comment)
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
	text += "\n\n"
	text += fmt.Sprintf("正片:\n%s\n", renderEpisodes(s.Episodes))
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
		text := fmt.Sprintf("%s+%d", tag.Name, tag.Count)
		if _, ok := wiki[tag.Name]; ok {
			text = fmt.Sprint(ui.Cyan(text))
		}
		arr = append(arr, text)
	}
	return "| " + strings.Join(arr, " | ") + " |"
}

func renderEpisodes(episodes *api.Episodes) string {
	if episodes == nil {
		return "NULL"
	}
	text := ""
	today := time.Now()
	for _, ep := range episodes.Data {
		if ep.Type != 0 {
			continue // Skip non-episode types
		}
		airTime, err := ep.GetAirTime()
		if err != nil {
			slog.Error("Failed to get air time for episode", "Error", err, "Episode", ep.ID)
			text += fmt.Sprintf("%d. %s %s\n", ep.Sort, ui.GraphicsColor(ep.GetName()), ui.Grey(ep.Airdate))
			continue
		}
		diff := dateCompare(airTime, today)
		if diff < 0 {
			// On aired
			text += fmt.Sprintf("%d. %s %s\n", ep.Sort, ep.GetName(), ui.Grey(ep.Airdate))
		} else if diff == 0 {
			// On airing today
			text += fmt.Sprintf("%d. %s %s\n", ep.Sort, ui.SecondaryText(ep.GetName()), ui.Grey(ep.Airdate))
		} else {
			// Not yet on aired
			text += fmt.Sprintf("%d. %s %s\n", ep.Sort, ui.GraphicsColor(ep.GetName()), ui.Grey(ep.Airdate))
		}
	}
	return text
}

// dateCompare returns the difference of a - b.
// Not exact result, only for comparing dates.
func dateCompare(a, b time.Time) int {
	if a.Year() != b.Year() {
		return a.Year() - b.Year()
	}
	return a.YearDay() - b.YearDay()
}
