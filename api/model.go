package api

import (
	"strings"
	"time"
)

type CollectionStatus string

func (c CollectionStatus) String() string {
	return string(c)
}

type (
	SubjectT      string
	EpisodeStatus string
)

var (
	C_STATUS = []CollectionStatus{Watching, Wish, Done, OnHold, Dropped}
	S_Type   = []SubjectT{Book, Anime, Music, Game, Real}
)

const (
	Wish     CollectionStatus = "wish"
	Done     CollectionStatus = "done"
	Watching CollectionStatus = "watching"
	OnHold   CollectionStatus = "stashed"
	Dropped  CollectionStatus = "dropped"
	All      CollectionStatus = "all" // custom
)

const (
	Book   SubjectT = "book"
	Anime  SubjectT = "anime"
	Music  SubjectT = "music"
	Game   SubjectT = "game"
	Real   SubjectT = "real"
	AllSub SubjectT = "all" // custom
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

var SubjectType map[string]int = map[string]int{
	"book":  1,
	"anime": 2,
	"music": 3,
	"game":  4,
	"real":  6,
}

var SubjectTypeRev map[int]string = map[int]string{
	1: "book",
	2: "anime",
	3: "music",
	4: "game",
	6: "real",
}

var EpisodeCollectionType map[string]int = map[string]int{
	"delete":  0,
	"wish":    1,
	"done":    2,
	"dropped": 3,
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
	return CollectionTypeRev[int(c.Type)]
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
	return TagNames(c.Subject.Tags)
}

type Subject struct {
	Images          map[string]string `json:"images"`
	Name            string            `json:"name"`
	NameCn          string            `json:"name_cn"`
	Summary         string            `json:"summary"`
	Nsfw            bool              `json:"nsfw"`
	Locked          bool              `json:"locked"`
	Tags            []Tag             `json:"tags"`
	Score           float64           `json:"score"`
	Type            uint32            `json:"type"`
	ID              uint32            `json:"id"`
	Eps             uint32            `json:"eps"`
	Volumes         uint32            `json:"volumes"`
	CollectionTotal uint32            `json:"collection_total"`
	Rank            uint32            `json:"rank"`
	Rating          Rating            `json:"rating"`
	CollectionCount CollectionCount   `json:"collection"`
	WikiTags        []string          `json:"meta_tags"` // tags from wiki users, not general user tags
	Platform        string            `json:"platform"`  // TV, Web, DLC, etc.
	// Optional fields
	Date    string           `json:"date"`    // can be empty
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

type UserEpisodeCollections struct {
	Total  int                     `json:"total"`
	Limit  int                     `json:"limit"`
	Offset int                     `json:"offset"`
	Data   []UserEpisodeCollection `json:"data"`
}

type UserEpisodeCollection struct {
	Episode Episode `json:"episode"`
	Type    int     `json:"type"`
}

type Episode struct {
	Airdate     string  `json:"airdate"`
	Name        string  `json:"name"`
	NameCn      string  `json:"name_cn"`
	Duration    string  `json:"duration"`
	Description string  `json:"description"`
	Ep          float32 `json:"ep"`
	SubjectId   int     `json:"subject_id"`
	Sort        float32 `json:"sort"`
	Comment     uint32  `json:"comment"`
	Id          int     `json:"id"`
	Type        int     `json:"type"`
	Disc        uint8   `json:"disc"`
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

func (s *Subject) GetName() string {
	if s.NameCn != "" {
		return s.NameCn
	}
	return s.Name
}

func (s *Subject) ToSlimSubject() SlimSubject {
	return SlimSubject{
		ID:              s.ID,
		Type:            s.Type,
		Name:            s.Name,
		NameCn:          s.NameCn,
		Images:          s.Images,
		ShortSummary:    s.Summary,
		Tags:            s.Tags,
		Score:           s.Score,
		Eps:             s.Eps,
		Volumes:         s.Volumes,
		CollectionTotal: s.CollectionTotal,
		Rank:            s.Rank,
		Date:            s.Date,
	}
}

func (s *Subject) GetAllTags() string {
	return TagNames(s.Tags)
}

// TagNames returns all users' tags as a space-separated string.
func TagNames(tags []Tag) string {
	if len(tags) == 0 {
		return ""
	}
	var names []string
	for _, tag := range tags {
		names = append(names, tag.Name)
	}
	return strings.Join(names, " ")
}
