package api

import (
	"strings"
	"time"
)

type CollectionStatus string

func (c CollectionStatus) String() string {
	return string(c)
}

type EpisodeStatus string

var C_STATUS = []CollectionStatus{Watching, Wish, Done, OnHold, Dropped}

const (
	Wish     CollectionStatus = "wish"
	Done     CollectionStatus = "done"
	Watching CollectionStatus = "watching"
	OnHold   CollectionStatus = "stashed"
	Dropped  CollectionStatus = "dropped"
	All      CollectionStatus = "all" // custom
)

var CollectionType = map[CollectionStatus]int{
	Wish:     1,
	Done:     2,
	Watching: 3,
	OnHold:   4,
	Dropped:  5,
}

var CollectionTypeRev = map[int]CollectionStatus{
	1: Wish,
	2: Done,
	3: Watching,
	4: OnHold,
	5: Dropped,
}

var SubjectMap map[string]int = map[string]int{
	"book":  1,
	"anime": 2,
	"music": 3,
	"game":  4,
	"real":  6,
}

var EpisodeCollectionType map[string]int = map[string]int{
	"delete":  0,
	"wish":    1,
	"done":    2,
	"dropped": 3,
}

var EpisodeType map[string]int = map[string]int{
	"DEFAULT": 0,
	"SP":      1,
	"OP":      2,
	"ED":      3,
}

var EpisodeTypeRev = map[int]string{
	0: "DEFAULT",
	1: "SP",
	2: "OP",
	3: "ED",
}

type UserCollections struct {
	Total  uint32                  `json:"total"`
	Limit  uint32                  `json:"limit"`
	Offset uint32                  `json:"offset"`
	Data   []UserSubjectCollection `json:"data"`
}

type UserSubjectCollection struct {
	UpdatedAt   time.Time   `json:"updated_at"`
	Comment     string      `json:"comment"`
	Tags        []string    `json:"tags"` // my tags
	Subject     SlimSubject `json:"subject"`
	SubjectID   uint32      `json:"subject_id"`
	VolStatus   uint32      `json:"vol_status"`
	EpStatus    uint32      `json:"ep_status"`    // current episode?
	SubjectType uint32      `json:"subject_type"` // 2: anime
	Type        uint32      `json:"type"`         // 3: watching
	Rate        uint32      `json:"rate"`
	Private     bool        `json:"private"`
}

// Return subject NameCn if available, otherwise return Name
func (c UserSubjectCollection) Name() string {
	if c.Subject.NameCn != "" {
		return c.Subject.NameCn
	}
	return c.Subject.Name
}

func (c *UserSubjectCollection) GetStatus() CollectionStatus {
	if status, ok := CollectionTypeRev[int(c.Type)]; ok {
		return status
	}
	return CollectionStatus("NULL")
}

func (c *UserSubjectCollection) SetStatus(status CollectionStatus) {
	c.Type = uint32(CollectionType[status])
}

func (c *UserSubjectCollection) GetSubjectType() string {
	return SubjectTypeRev[int(c.SubjectType)]
}

func (c *UserSubjectCollection) GetTags() string {
	return strings.Join(c.Tags, " ")
}

func (c *UserSubjectCollection) GetAllTags() string {
	return tagNames(c.Subject.Tags)
}

type Subject struct {
	SlimSubject
	Images          map[string]string `json:"images"`
	Summary         string            `json:"summary"`
	Nsfw            bool              `json:"nsfw"`
	Series          bool              `json:"series"` // Is the main subject for books or not
	Locked          bool              `json:"locked"`
	Score           float64           `json:"score"`
	TotalEps        uint32            `json:"total_episodes"`
	Rank            uint32            `json:"rank"`
	Rating          Rating            `json:"rating"`
	CollectionCount CollectionCount   `json:"collection"`
	WikiTags        []string          `json:"meta_tags"` // tags from wiki users, not general user tags
	Platform        string            `json:"platform"`  // TV, Web, DLC, etc.
	// Optional fields
	InfoBox []map[string]any `json:"infobox"` // A list of ordered maps for additional information
}

type SlimSubject struct {
	ID              uint32            `json:"id"`
	Type            uint32            `json:"type"`
	Images          map[string]string `json:"images"`
	Name            string            `json:"name"`
	NameCn          string            `json:"name_cn"`
	ShortSummary    string            `json:"short_summary"`
	Tags            []Tag             `json:"tags"` // frist 10 tags
	Score           float64           `json:"score"`
	Eps             uint32            `json:"eps"`
	Volumes         uint32            `json:"volumes"`
	CollectionTotal uint32            `json:"collection_total"`
	Rank            uint32            `json:"rank"`
	Date            string            `json:"date"` // can be empty
}

func (s *SlimSubject) GetName() string {
	if s.NameCn != "" {
		return s.NameCn
	}
	return s.Name
}

func (s *SlimSubject) GetAllTags() string {
	return tagNames(s.Tags)
}

// Get first num tags in a space-separated string
func (s *SlimSubject) GetTags(num int) string {
	if num <= 0 || len(s.Tags) == 0 {
		return ""
	}
	if num > len(s.Tags) {
		num = len(s.Tags)
	}
	names := make([]string, num)
	for i := 0; i < num; i++ {
		names[i] = s.Tags[i].Name
	}
	return strings.Join(names, " ")
}

// Returned type of /v0/users/-/collections/{subject_id}/episodes
type UserEpisodeCollections struct {
	Total  int                     `json:"total"`
	Limit  int                     `json:"limit"`
	Offset int                     `json:"offset"`
	Data   []UserEpisodeCollection `json:"data"`
}

func (e *UserEpisodeCollections) Status() {
	panic("Not impelemented")
}

// User watched episode. The latest.
func (e *UserEpisodeCollections) Latest() *Episode {
	doneType := EpisodeCollectionType["done"]
	for i := len(e.Data); i >= 0; i -= 1 {
		userEpisode := e.Data[i]
		if userEpisode.Type != doneType {
			return &userEpisode.Episode
		}
	}
	return nil
}

type UserEpisodeCollection struct {
	Episode Episode `json:"episode"`
	Type    int     `json:"type"`
}

// Subject's episode information. Not users'.
type Episodes struct {
	Total  int       `json:"total"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
	Data   []Episode `json:"data"`
}

func (e *Episodes) Status() {
	panic("Not impelemented")
}

// Latest on aired episode Sort
func (e *Episodes) Latest() int {
	today := time.Now()
	// TODO: When the episode list is too long, the data may not contain the latest episode.

	// FIXME: binary search
	for i := len(e.Data) - 1; i >= 0; i -= 1 {
		parsed, err := parseDate(e.Data[i].Airdate)
		if err != nil {
			return -1
		}
		if parsed.Before(today) {
			return e.Data[i].Sort
		}
	}
	return -1
}

type Episode struct {
	Airdate     string `json:"airdate"`
	Name        string `json:"name"`
	NameCn      string `json:"name_cn"`
	Duration    string `json:"duration"`
	Description string `json:"description"`
	Ep          int    `json:"ep"`   // Episode number of current season
	Sort        int    `json:"sort"` // Episode number of all seasons
	SubjectId   int    `json:"subject_id"`
	Comment     uint32 `json:"comment"`
	ID          int    `json:"id"`
	Type        int    `json:"type"`
	Disc        uint8  `json:"disc"`
}

func (e *Episode) GetName() string {
	if e.NameCn != "" {
		return e.NameCn
	} else if e.Name != "" {
		return e.Name
	}
	return "N/A"
}

func (e *Episode) OnAirToday() bool {
	today := time.Now()
	today = today.Add(time.Hour * 9) // TO JST
	todayStr := today.Format("2006-01-02")
	return e.Airdate == todayStr
}

func (e *Episode) GetAirTime() (time.Time, error) {
	return parseDate(e.Airdate)
}

type Tag struct {
	Name  string
	Count int
}

type Rating struct {
	Rank  int     `json:"rank"`
	Total int     `json:"total"`
	Count Count   `json:"count"`
	Score float64 `json:"score"`
}

type Count struct {
	Field1  uint32
	Field2  uint32
	Field3  uint32
	Field4  uint32
	Field5  uint32
	Field6  uint32
	Field7  uint32
	Field8  uint32
	Field9  uint32
	Field10 uint32
}

type Calendar struct {
	Weekday Weekday        `json:"weekday"`
	Items   []CalendarItem `json:"items"`
}

// CalenderItem uses the old subject type so not 100% accurate
type CalendarItem struct {
	ID              int             `json:"id"`
	URL             string          `json:"url"`
	Summary         string          `json:"summary"`
	AirWeekday      int             `json:"air_weekday"`
	CollectionCount CollectionCount `json:"collection"`
	EpsCount        int             `json:"eps_count"`
	SlimSubject
}

type CollectionCount struct {
	Watching uint32 `json:"doing"`
	Wish     uint32 `json:"wish"`
	Done     uint32 `json:"collect"`
	OnHold   uint32 `json:"on_hold"`
	Dropped  uint32 `json:"dropped"`
}

type Weekday struct {
	ID uint32 `json:"id"`
	EN string `json:"en"`
	CN string `json:"cn"`
	JA string `json:"ja"`
}

// tagNames returns all users' tags as a space-separated string.
func tagNames(tags []Tag) string {
	if len(tags) == 0 {
		return ""
	}
	var names []string
	for _, tag := range tags {
		names = append(names, tag.Name)
	}
	return strings.Join(names, " ")
}

// parseDate parses a date string in the format "2006-01-02"
func parseDate(dateString string) (time.Time, error) {
	parsedTime, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		return time.Now(), err
	}
	return parsedTime, nil
}
