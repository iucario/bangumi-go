package api

type SubjectType int

func (s SubjectType) String() string {
	if r, ok := SubjectTypeRev[int(s)]; ok {
		return r
	}
	return ""
}

func (s SubjectType) CN() string {
	switch s {
	case BOOK:
		return "书籍"
	case ANIME:
		return "动画"
	case MUSIC:
		return "音乐"
	case GAME:
		return "游戏"
	case REAL:
		return "三次元"
	default:
		return ""
	}
}

var (
	BOOK  = SubjectType(1)
	ANIME = SubjectType(2)
	MUSIC = SubjectType(3)
	GAME  = SubjectType(4)
	REAL  = SubjectType(6)
)

var S_TYPE_ALL = []SubjectType{BOOK, ANIME, MUSIC, GAME, REAL}

var SubjectTypeMap = map[string]SubjectType{
	"book":  BOOK,
	"anime": ANIME,
	"music": MUSIC,
	"game":  GAME,
	"real":  REAL,
	"all":   SubjectType(0), // Custom type for all subjects
}

var SubjectTypeRev map[int]string = map[int]string{
	1: "book",
	2: "anime",
	3: "music",
	4: "game",
	6: "real",
}

type Sort string

var (
	MATCH = Sort("match")
	HEAT  = Sort("heat")
	RANK  = Sort("rank")
	SCORE = Sort("score")
)

var SortMap = map[string]Sort{
	"match": MATCH,
	"heat":  HEAT,
	"rank":  RANK,
	"score": SCORE,
}

type Payload struct {
	Keyword string `json:"keyword"`
	Sort    Sort   `json:"sort"` // match, heat, rank, score
	Filter  Filter `json:"filter"`
}

// All lists are AND relation execpt for Type
type Filter struct {
	Type     []SubjectType `json:"type,omitempty"`      // OR relation
	MetaTags []string      `json:"meta_tags,omitempty"` // AND relation. add a '-' at the beginning to exclude
	Tag      []string      `json:"tag,omitempty"`       // AND relation. '-' to exclude
	AirDate  []string      `json:"air_date,omitempty"`  // AND relation. YYYY-MM-DD. E.g. [">=2020-07-01", "<2020-10-01"]
	Rating   []string      `json:"rating,omitempty"`    // AND. [">=6", "<8"]
	Rank     []string      `json:"rank,omitempty"`      // AND [">10", "<=18"]
	NSFW     bool          `json:"nsfw,omitempty"`      // True for NSFW only, False SFW. Null for both
}

type SubjectList struct {
	Total  int       `json:"total"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
	Data   []Subject `json:"data"`
}
