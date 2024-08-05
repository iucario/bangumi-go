package list

import (
	"fmt"
	"log"

	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd"
	"github.com/iucario/bangumi-go/cmd/auth"
	"github.com/spf13/cobra"
)

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
func ListUserCollection(access_token string, username string, subjectType string, collectionType string, limit int, offset int) (api.UserCollections, error) {
	userCollections := api.UserCollections{}
	var subjectTypeInt int
	var collectionTypeInt int
	if subjectType == "all" {
		subjectTypeInt = -1
	} else {
		subjectTypeInt = api.SubjectType[subjectType]
	}
	if collectionType == "all" {
		collectionTypeInt = -1
	} else {
		collectionTypeInt = api.CollectionType[collectionType]
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
func ListCollection(access_token string, username string, subjectType int, collectionType int, limit int, offset int) (api.UserCollections, error) {

	baseUrl := fmt.Sprintf("https://api.bgm.tv/v0/users/%s/collections", username)
	var url string
	switch {
	case subjectType == -1 && collectionType == -1:
		url = fmt.Sprintf("%s?limit=%d&offset=%d", baseUrl, limit, offset)
	case subjectType == -1:
		url = fmt.Sprintf("%s?type=%d&limit=%d&offset=%d", baseUrl, collectionType, limit, offset)
	case collectionType == -1:
		url = fmt.Sprintf("%s?subject_type=%d&limit=%d&offset=%d", baseUrl, subjectType, limit, offset)
	default:
		url = fmt.Sprintf("%s?subject_type=%d&type=%d&limit=%d&offset=%d", baseUrl, subjectType, collectionType, limit, offset)
	}

	log.Printf("ListCollection api: %s\n", url)

	userCollections := api.UserCollections{}
	err := api.AuthenticatedGetRequest(url, access_token, &userCollections)
	return userCollections, err
}
