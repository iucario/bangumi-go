package api

type SubjectType int

var (
	BOOK  = SubjectType(1)
	ANIME = SubjectType(2)
	MUSIC = SubjectType(3)
	GAME  = SubjectType(4)
	REAL  = SubjectType(6)
)

type Sort string

var (
	MATCH = Sort("match")
	HEAT  = Sort("heat")
	RANK  = Sort("rank")
	SCORE = Sort("score")
)

type Payload struct {
	Keyword string `json:"keyword"`
	Sort    Sort   `json:"sort"` // match, heat, rank, score
	Filter  Filter `json:"filter"`
}

// All lists are AND relation.
type Filter struct {
	Type     []SubjectType `json:"type,omitempty"`
	MetaTags []string      `json:"meta_tags,omitempty"` // AND relation. add a '-' at the beginning to exclude
	Tag      []string      `json:"tag,omitempty"`       // AND relation. '-' to exclude
	AirDate  string        `json:"air_date,omitempty"`  // AND relation. YYYY-MM-DD. E.g. [">=2020-07-01", "<2020-10-01"]
	Rating   int           `json:"rating,omitempty"`    // AND. [">=6", "<8"]
	Rank     int           `json:"rank,omitempty"`      // AND [">10", "<=18"]
	NSFW     bool          `json:"nsfw,omitempty"`      // True for NSFW only, False SFW. Null for both
}

type SubjectList struct {
	Total  int       `json:"total"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
	Data   []Subject `json:"data"`
}
