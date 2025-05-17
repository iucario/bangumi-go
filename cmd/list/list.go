package list

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List collection",
	Run: func(cmd *cobra.Command, args []string) {
		subjectType, _ := cmd.Flags().GetString("subject")
		collectionType, _ := cmd.Flags().GetString("collection")

		if api.CollectionStatus(collectionType) == "" {
			fmt.Printf("Invalid collection type: %s\n", collectionType)
		}

		authClient := api.NewAuthClientWithConfig()
		user := api.NewUser(authClient)
		userInfo, err := user.GetUserInfo()
		api.AbortOnError(err)

		options := UserListOptions{
			SubjectType:    subjectType,
			Username:       userInfo.Username,
			CollectionType: api.CollectionStatus(collectionType),
			Limit:          30,
			Offset:         0,
		}
		watchCollections, _ := ListUserCollection(authClient, options)
		slog.Info(fmt.Sprintf("collections in watching: %d\n", watchCollections.Total))

		fmt.Printf("Total: %d. Showing: %d\n", watchCollections.Total, len(watchCollections.Data))
		for i, collection := range watchCollections.Data {
			name := collection.Subject.NameCn
			if name == "" {
				name = collection.Subject.Name
			}
			fmt.Printf("%d. %d/%d %s\n", i+1, collection.EpStatus, collection.Subject.Eps, name)
		}
	},
}

type UserListOptions struct {
	Username       string
	SubjectType    string
	CollectionType api.CollectionStatus
	Limit          int
	Offset         int
}

// subjectType: "anime", "real", "all".
// collectionType: "wish", "done", "watch", "on_hold", "dropped", "all".
func ListUserCollection(authClient *api.AuthClient, options UserListOptions) (*api.UserCollections, error) {
	var subjectTypeInt int
	var collectionTypeInt int
	if options.SubjectType == "all" {
		subjectTypeInt = -1
	} else {
		subjectTypeInt = api.SubjectType[options.SubjectType]
	}
	if options.CollectionType == "all" {
		collectionTypeInt = -1
	} else {
		collectionTypeInt = api.CollectionType[options.CollectionType]
	}

	params := ListParams{
		Username:       options.Username,
		SubjectType:    subjectTypeInt,
		CollectionType: collectionTypeInt,
		Limit:          options.Limit,
		Offset:         options.Offset,
	}

	return ListCollection(authClient, params)
}

type ListParams struct {
	Username       string `json:"username"`
	SubjectType    int    `json:"subject_type"`
	CollectionType int    `json:"type"`
	Limit          int    `json:"limit"`
	Offset         int    `json:"offset"`
}

// List user bangumi collection
// Does not include subjectType or collectionType in parameters if is set to -1.
func ListCollection(authClient *api.AuthClient, params ListParams) (*api.UserCollections, error) {
	baseUrl := fmt.Sprintf("https://api.bgm.tv/v0/users/%s/collections", params.Username)
	url, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}
	queries := url.Query()
	queries.Set("limit", fmt.Sprintf("%d", params.Limit))
	queries.Set("offset", fmt.Sprintf("%d", params.Offset))
	switch {
	case params.SubjectType == -1 && params.CollectionType == -1:
		break
	case params.SubjectType == -1:
		queries.Set("type", fmt.Sprintf("%d", params.CollectionType))
	case params.CollectionType == -1:
		queries.Set("subject_type", fmt.Sprintf("%d", params.SubjectType))
	default:
		queries.Set("type", fmt.Sprintf("%d", params.CollectionType))
		queries.Set("subject_type", fmt.Sprintf("%d", params.SubjectType))
	}
	url.RawQuery = queries.Encode()

	slog.Info(fmt.Sprintf("ListCollection: %s", url.String()))

	b, err := authClient.Get(url.String())
	if err != nil {
		return nil, err
	}
	userCollections := api.UserCollections{}
	if err := json.Unmarshal(b, &userCollections); err != nil {
		return nil, err
	}
	return &userCollections, err
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
