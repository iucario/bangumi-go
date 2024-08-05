package list

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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
		subjectType, _ := cmd.Flags().GetString("subject")
		collectionType, _ := cmd.Flags().GetString("collection")
		credential, _ := auth.LoadCredential()
		userInfo, err := auth.GetUserInfo(credential.AccessToken)
		auth.Check(err)
		watchCollections, _ := ListUserCollection(credential.AccessToken, userInfo.Username, subjectType, collectionType, 30, 0)
		log.Printf("collections in watching: %d\n", watchCollections.Total)

		fmt.Printf("Total: %d. Showing: %d\n", watchCollections.Total, len(watchCollections.Data))
		for i, collection := range watchCollections.Data {
			fmt.Printf("%d. %d/%d %s\n", i+1, collection.EpStatus, collection.Subject.Eps, collection.Subject.NameCn)
		}
	},
}

func init() {
	var subjectType string
	var collectionType string
	listCmd.Flags().StringVarP(&collectionType, "collection", "c", "watch",
		"Collection type: wish, done, watch, onhold, dropped, all.")
	listCmd.Flags().StringVarP(&subjectType, "subject", "s", "all",
		"Subject type: book, anime, music, game, real, all.")
	cmd.RootCmd.AddCommand(listCmd)
}

// subjectType: "anime", "real", "all".
// collectionType: "wish", "done", "watch", "on_hold", "dropped", "all".
func ListUserCollection(access_token string, username string, subjectType string, collectionType string, limit int, offset int) (UserCollections, error) {
	userCollections := UserCollections{}
	var subjectTypeInt int
	var collectionTypeInt int
	if subjectType == "all" {
		subjectTypeInt = -1
	} else {
		subjectTypeInt = SubjectType[subjectType]
	}
	if collectionType == "all" {
		collectionTypeInt = -1
	} else {
		collectionTypeInt = CollectionType[collectionType]
	}

	credential, err := auth.LoadCredential()
	if err != nil {
		return userCollections, err
	}

	userCollections, err = ListCollection(credential.AccessToken, username, subjectTypeInt, collectionTypeInt, limit, offset)
	if err != nil {
		return userCollections, err
	}

	return userCollections, nil
}

// List user bangumi collection
// Does not include subjectType or collectionType in parameters if is set to -1.
func ListCollection(access_token string, username string, subjectType int, collectionType int, limit int, offset int) (UserCollections, error) {
	userCollections := UserCollections{}
	baseUrl := fmt.Sprintf("https://api.bgm.tv/v0/users/%s/collections", username)
	var api string
	switch {
	case subjectType == -1 && collectionType == -1:
		api = fmt.Sprintf("%s?limit=%d&offset=%d", baseUrl, limit, offset)
	case subjectType == -1:
		api = fmt.Sprintf("%s?type=%d&limit=%d&offset=%d", baseUrl, collectionType, limit, offset)
	case collectionType == -1:
		api = fmt.Sprintf("%s?subject_type=%d&limit=%d&offset=%d", baseUrl, subjectType, limit, offset)
	default:
		api = fmt.Sprintf("%s?subject_type=%d&type=%d&limit=%d&offset=%d", baseUrl, subjectType, collectionType, limit, offset)
	}

	log.Printf("ListCollection api: %s\n", api)

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
