package subject

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd"
	"github.com/spf13/cobra"
)

var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "Subject actions",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`Available commands:
bgm sub info <subject_id>
bgm sub edit <subject_id> [-w <episode number>]`)
	},
}

func init() {
	cmd.RootCmd.AddCommand(subCmd)
}

func GetUserSubjectCollection(token string, username string, subjectId int) (api.UserSubjectCollection, error) {
	url := fmt.Sprintf("https://api.bgm.tv/v0/users/%s/collections/%d", username, subjectId)
	userSubjectCollection := api.UserSubjectCollection{}
	err := api.AuthenticatedGetRequest(url, token, &userSubjectCollection)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get user subject collection: %v\n", err))
	}
	return userSubjectCollection, err
}

func GetUserEpisodeCollections(token, username string, subjectId, offset, limit, episode_type int) (api.UserEpisodeCollections, error) {
	url := fmt.Sprintf("https://api.bgm.tv/v0/users/-/collections/%d/episodes", subjectId)
	userEpisodeCollections := api.UserEpisodeCollections{}
	err := api.AuthenticatedGetRequest(url, token, &userEpisodeCollections)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get user episode collection: %v\n", err))
	}
	return userEpisodeCollections, err
}

// status: delete, wish, done, dropped
func PutEpisode(token string, episodeId int, status string) error {
	url := fmt.Sprintf("https://api.bgm.tv/v0/users/-/collections/-/episodes/%d", episodeId)
	typeInt := api.EpisodeCollectionType[status]
	jsonBytes := []byte(fmt.Sprintf(`{
		"type": %d}
	`, typeInt))
	err := api.PutRequest(url, token, jsonBytes, nil)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to put episode: %v", err))
	}
	return err
}

// status: delete, wish, done, dropped
func PatchEpisodes(token string, subjectId int, episodeIds []int, status string) error {
	url := fmt.Sprintf("https://api.bgm.tv/v0/users/-/collections/%d/episodes", subjectId)
	slog.Info(fmt.Sprintf("PATCH to %v", url))
	typeInt := api.EpisodeCollectionType[status]
	requestBody := struct {
		EpisodeID []int `json:"episode_id"`
		Type      int   `json:"type"`
	}{
		EpisodeID: episodeIds,
		Type:      typeInt,
	}
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to marshal request body: %v", err))
		return err
	}
	err = api.PatchRequest(url, token, jsonBytes, nil)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to patch episodes: %v", err))
	}
	return err
}
