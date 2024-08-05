package list

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/iucario/bangumi-go/cmd"
	"github.com/iucario/bangumi-go/cmd/auth"
	"github.com/spf13/cobra"
)

var CollectionType map[string]int = map[string]int{
	"wish":    1,
	"done":    2,
	"watch":   3,
	"on_hold": 4,
	"dropped": 5,
}

var SubjectType map[string]int = map[string]int{
	"book":  1,
	"anime": 2,
	"music": 3,
	"game":  4,
	"real":  6,
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
	Rating          map[string]uint32 `json:"rating"`
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

type Tag struct {
	Name  string
	Count int
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List collection",
	Run: func(cmd *cobra.Command, args []string) {
		credential, _ := auth.LoadCredential()
		userInfo, err := auth.GetUserInfo(credential.AccessToken)
		auth.Check(err)
		userCollections, _ := ListAnimeCollection(credential.AccessToken, userInfo.Username, 3, 20, 0)
		for i, collection := range userCollections.Data {
			fmt.Printf("%d. %s\n", i+1, collection.Subject.Name)
		}
	},
}

func ListAnimeCollection(access_token string, username string, collectionType int, limit int, offset int) (UserCollections, error) {
	userCollections := UserCollections{}
	subjectType := SubjectType["anime"]

	credential, err := auth.LoadCredential()
	if err != nil {
		return userCollections, err
	}

	userCollections, err = ListCollection(credential.AccessToken, username, subjectType, collectionType, limit, offset)
	if err != nil {
		return userCollections, err
	}

	return userCollections, nil
}

// List user bangumi collection
func ListCollection(access_token string, username string, subjectType int, collectionType int, limit int, offset int) (UserCollections, error) {
	userCollections := UserCollections{}
	api := fmt.Sprintf("https://api.bgm.tv/v0/users/%s/collections?subject_type=%d&type=%d&limit=%d&offset=%d",
		username, subjectType, collectionType, limit, offset)
	req, err := http.NewRequest("GET", api, nil)
	req.Header.Set("User-Agent", auth.UserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", access_token))
	if err != nil {
		return userCollections, err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return userCollections, err
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)
	bodyString := string(bodyBytes)
	if res.StatusCode != 200 {
		return userCollections, fmt.Errorf("[error] status code: %d, response: %s", res.StatusCode, bodyString)
	}

	err = json.Unmarshal(bodyBytes, &userCollections)
	if err != nil {
		return userCollections, err
	}
	return userCollections, nil
}

func init() {
	cmd.RootCmd.AddCommand(listCmd)
}
