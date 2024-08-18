package subject

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd/auth"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Subject information",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		watch, _ := cmd.Flags().GetInt("watch")
		subjectId, err := strconv.Atoi(id)
		if err != nil {
			slog.Error(fmt.Sprintf("Invalid subject ID: %s", id))
			return
		}
		slog.Info(fmt.Sprintf("edit subjectId=%d watch=%d\n", subjectId, watch))
		credential, _ := auth.LoadCredential()
		userInfo, err := auth.GetUserInfo(credential.AccessToken)
		auth.AbortOnError(err)

		if watch == -1 {
			WatchNextEpisode(credential.AccessToken, userInfo.Username, subjectId)
		} else {
			WatchToEpisode(credential.AccessToken, userInfo.Username, subjectId, watch)
		}
	},
}

func init() {
	var watch int
	editCmd.Flags().IntVarP(&watch, "watch", "w", -1, "Watch to episode [n]. -1 for next episode.")
	subCmd.AddCommand(editCmd)
}

func WatchNextEpisode(token string, username string, subjectId int) {
	userSubjectCollection, err := GetUserSubjectCollection(token, username, subjectId)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get user subject collection: %v\n", err))
	}
	epStatus := userSubjectCollection.EpStatus
	totalEps := userSubjectCollection.Subject.Eps
	subjectName := userSubjectCollection.Subject.NameCn
	if epStatus > totalEps {
		slog.Error(fmt.Sprintf("No more episodes to watch. Current: %d, Total: %d\n", epStatus, totalEps))
	}

	userEpisodeCollections, err := GetUserEpisodeCollections(token, username, subjectId, 0, 100, 0)
	episode, err := getCurrentEpisode(userEpisodeCollections.Data)
	if err != nil {
		slog.Error(fmt.Sprintf("%v\n", err))
	}

	PutEpisode(token, int(episode.Episode.Id), "done")
	epName := episode.Episode.NameCn
	slog.Info(fmt.Sprintf("Marked as done: %s episode %d. %s\n", subjectName, episode.Episode.Id, epName))
}

// Mark 1 to n episodes as done, the rest as delete
func WatchToEpisode(token string, username string, subjectId int, episodeNum int) {
	userEpisodeCollections, err := GetUserEpisodeCollections(token, username, subjectId, 0, 100, 0)
	if err != nil {
		slog.Error(fmt.Sprintf("%v\n", err))
	}
	totalEps := len(userEpisodeCollections.Data)
	if episodeNum > totalEps {
		slog.Warn(fmt.Sprintf("Episode number %d exceeds total episodes: %d. Marking all.\n", episodeNum, totalEps))
	}
	validNum := max(0, min(episodeNum, totalEps))
	watchList := make([]int, validNum)
	deleteList := make([]int, totalEps-validNum)
	for i, userEpisode := range userEpisodeCollections.Data {
		if i < validNum {
			watchList[i] = int(userEpisode.Episode.Id)
		} else {
			deleteList[i-validNum] = int(userEpisode.Episode.Id)
		}
	}

	PatchEpisodes(token, subjectId, watchList, "done")
	PatchEpisodes(token, subjectId, deleteList, "delete")

}

// Return the first episode that is not done
func getCurrentEpisode(userEpisodeCollection []api.UserEpisodeCollection) (api.UserEpisodeCollection, error) {

	doneType := api.EpisodeCollectionType["done"]
	for _, userEpisode := range userEpisodeCollection {
		if userEpisode.Type != doneType {
			return userEpisode, nil
		}
	}
	return userEpisodeCollection[0], fmt.Errorf("No more episodes to watch")
}
