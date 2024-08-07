package api

import "time"

var CollectionType map[string]int = map[string]int{
	"wish":    1,
	"done":    2,
	"watch":   3,
	"onhold":  4,
	"dropped": 5,
}

var SubjectType map[string]int = map[string]int{
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
	Collection      map[string]uint32 `json:"collection"`
	Date            string            `json:"date"` // can be empty
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
