package subject

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"

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

func GetUserSubjectCollection(c *api.AuthClient, username string, subjectId int) (api.UserSubjectCollection, error) {
	url := fmt.Sprintf("https://api.bgm.tv/v0/users/%s/collections/%d", username, subjectId)
	b, err := c.Get(url)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get user subject collection: %v\n", err))
	}
	userSubjectCollection := api.UserSubjectCollection{}
	err = json.Unmarshal(b, &userSubjectCollection)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to unmarshal user subject collection: %v\n", err))
	}
	return userSubjectCollection, err
}

// Get user's episode info
func GetUserEpisodeCollections(c *api.AuthClient, subjectId, offset, limit, episode_type int) (api.UserEpisodeCollections, error) {
	url := fmt.Sprintf("https://api.bgm.tv/v0/users/-/collections/%d/episodes", subjectId)
	b, err := c.Get(url)
	if err != nil {
		slog.Error("getting user episode collection", "Error", err)
	}
	userEpisodeCollections := api.UserEpisodeCollections{}
	err = json.Unmarshal(b, &userEpisodeCollections)
	if err != nil {
		slog.Error("unmarshalling user episode collection", "Error", err)
	}
	return userEpisodeCollections, err
}

// Get subject's episodes info, not user's
func GetEpisodes(c *api.HTTPClient, subjectID, offest, limit int) (*api.Episodes, error) {
	url := fmt.Sprintf("https://api.bgm.tv/v0/episodes?subject_id=%d&offset=%d&limit=%d", subjectID, offest, limit)
	b, err := c.Get(url)
	if err != nil {
		slog.Error("getting episodes", "Error", err)
	}
	episodes := api.Episodes{}
	err = json.Unmarshal(b, &episodes)
	if err != nil {
		slog.Error("unmarshalling episodes", "Error", err)
	}
	return &episodes, err
}

// status: wish, done, watch, onhold, dropped
// ep_status and vol_status are only used for book
func PostCollection(c *api.AuthClient, subjectId int, status api.CollectionStatus, tags []string, comment string, rate int, private bool) error {
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
	_, err = c.Post(url, jsonBytes)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to post collection: %v", err))
		return err
	}
	slog.Debug(fmt.Sprintf("Successfully set subject %d to %s", subjectId, status))
	return nil // TODO: maybe return the collection struct
}

type PatchCollectionParams struct {
	Status    *api.CollectionStatus
	Tags      *[]string
	Comment   *string
	Rate      *int
	Private   *bool
	EpStatus  *int // not used
	VolStatus *int // not used
}

// PatchCollection only patches changed information. All arguments except c and subjectId are optional (can be nil/zero value).
// Patching episode status can only be used for books, not animes.
// PatchCollection is not used anywhere because it can be replaced by PostCollection
func PatchCollection(c *api.AuthClient, original, updated *api.UserSubjectCollection) error {
	url := fmt.Sprintf("https://api.bgm.tv/v0/users/-/collections/%d", updated.SubjectID)
	requestBody := make(map[string]any)
	// Compare and patch only changed fields (excluding episode/volume status)
	if updated.GetStatus() != original.GetStatus() {
		requestBody["type"] = api.CollectionType[updated.GetStatus()]
	}
	if updated.Rate != original.Rate {
		requestBody["rate"] = int(updated.Rate)
	}
	if updated.Comment != original.Comment {
		requestBody["comment"] = updated.Comment
	}
	if updated.Private != original.Private {
		requestBody["private"] = updated.Private
	}
	if !slices.Equal(updated.Tags, original.Tags) {
		requestBody["tags"] = updated.Tags
	}
	if len(requestBody) == 0 {
		slog.Warn("No fields to patch in PatchCollection")
		return nil
	}
	jsonBytes, err := json.Marshal(requestBody)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to marshal request body: %v", err))
		return err
	}
	_, err = c.Patch(url, jsonBytes)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to patch collection: %v", err))
		return err
	}
	slog.Debug("Successfully patched subject", "ID", updated.SubjectID)
	return nil
}

// status: delete, wish, done, dropped
func PutEpisode(c *api.AuthClient, episodeId int, status string) error {
	url := fmt.Sprintf("https://api.bgm.tv/v0/users/-/collections/-/episodes/%d", episodeId)
	typeInt := api.EpisodeCollectionType[status]
	data := struct {
		Type int `json:"type"`
	}{
		Type: typeInt,
	}
	jsonBytes, _ := json.Marshal(data)
	_, err := c.Put(url, jsonBytes)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to put episode: %v", err))
	}

	return err
}

// status: delete, wish, done, dropped
func PatchEpisodes(c *api.AuthClient, subjectId int, episodeIds []int, status string) error {
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
	_, err = c.Patch(url, jsonBytes)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to patch episodes: %v", err))
	}
	return err
}

// CollectionInfoChanged returns true if any of the main collection info fields have changed.
func CollectionInfoChanged(original, updated *api.UserSubjectCollection) bool {
	if original.GetStatus() != updated.GetStatus() {
		return true
	}
	if !slices.Equal(original.Tags, updated.Tags) {
		return true
	}
	if original.Comment != updated.Comment {
		return true
	}
	if original.Rate != updated.Rate {
		return true
	}
	if original.Private != updated.Private {
		return true
	}
	return false
}
