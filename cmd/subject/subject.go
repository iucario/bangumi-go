package subject

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd"
	"github.com/spf13/cobra"
)

var ConfigDir string

var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "Subject actions",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(`Available commands:
bgm sub info <subject_id>
bgm sub status <subject_id>
bgm sub edit <subject_id> [-w <episode number>]`)
	},
}

func init() {
	cmd.RootCmd.AddCommand(subCmd)
	ConfigDir = cmd.ConfigDir
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

func GetUserEpisodeCollections(token string, subjectId, offset, limit, episode_type int) (api.UserEpisodeCollections, error) {
	url := fmt.Sprintf("https://api.bgm.tv/v0/users/-/collections/%d/episodes", subjectId)
	userEpisodeCollections := api.UserEpisodeCollections{}
	err := api.AuthenticatedGetRequest(url, token, &userEpisodeCollections)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get user episode collection: %v\n", err))
	}
	return userEpisodeCollections, err
}

// status: wish, done, watch, onhold, dropped
// ep_status and vol_status are only used for book
func PostCollection(token string, subjectId int, status api.CollectionStatus, tags []string, comment string, rate int, private bool) error {
	url := fmt.Sprintf("https://api.bgm.tv/v0/users/-/collections/%d", subjectId)
	requestBody := struct {
		Type    int      `json:"type"`
		Rate    int      `json:"rate"`
		Comment string   `json:"comment"`
		Private bool     `json:"private"`
		Tags    []string `json:"tags"`
	}{
		Type:    api.CollectionType[status],
		Rate:    rate,
		Comment: comment,
		Private: private,
		Tags:    tags,
	}
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to marshal request body: %v", err))
		return err
	}
	err = api.PostRequest(url, token, jsonBytes, nil)
	if err != nil {
	} else {
		slog.Info(fmt.Sprintf("Successfully set subject %d to %s", subjectId, status))
	}
	return err
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
	slog.Info(fmt.Sprintf("PATCH status %s to subject %d", status, subjectId))
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
