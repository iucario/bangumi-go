package subject

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/iucario/bangumi-go/api"
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
		authClient := api.NewAuthClientWithConfig()
		user := api.NewUser(authClient)
		userInfo, err := user.GetUserInfo()
		api.AbortOnError(err)

		if watch == -1 {
			WatchNextEpisode(authClient, userInfo.Username, subjectId)
		} else {
			WatchToEpisode(authClient, subjectId, watch)
		}
	},
}

func init() {
	var watch int
	editCmd.Flags().IntVarP(&watch, "watch", "w", -1, "Watch to episode [n]. -1 for next episode.")
	subCmd.AddCommand(editCmd)
}

func WatchNextEpisode(c *api.AuthClient, username string, subjectId int) {
	userSubjectCollection, err := GetUserSubjectCollection(c, username, subjectId)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get user subject collection: %v\n", err))
	}
	epStatus := userSubjectCollection.EpStatus
	totalEps := userSubjectCollection.Subject.Eps
	subjectName := userSubjectCollection.Subject.NameCn
	if epStatus > totalEps {
		slog.Error(fmt.Sprintf("No more episodes to watch. Current: %d, Total: %d\n", epStatus, totalEps))
	}

	userEpisodeCollections, err := GetUserEpisodeCollections(c, subjectId, 0, 100, 0)
	if err != nil {
		slog.Error(fmt.Sprintf("%v\n", err))
	}
	episode, err := getCurrentEpisode(userEpisodeCollections.Data)
	if err != nil {
		slog.Error(fmt.Sprintf("%v\n", err))
	}

	err = PutEpisode(c, int(episode.Episode.ID), "done")
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to mark episode as done: %v\n", err))
		return
	}
	epName := episode.Episode.NameCn
	slog.Info(fmt.Sprintf("Marked as done: %s episode %d. %s\n", subjectName, episode.Episode.ID, epName))
}

// Mark 1 to n episodes as done, the rest as delete
func WatchToEpisode(c *api.AuthClient, subjectId int, episodeNum int) {
	userEpisodeCollections, err := GetUserEpisodeCollections(c, subjectId, 0, 100, 0)
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
			watchList[i] = int(userEpisode.Episode.ID)
		} else {
			deleteList[i-validNum] = int(userEpisode.Episode.ID)
		}
	}

	err = PatchEpisodes(c, subjectId, watchList, "done")
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to mark episodes as done: %v\n", err))
		return
	}
	err = PatchEpisodes(c, subjectId, deleteList, "delete")
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to delete episodes: %v\n", err))
		return
	}
}

// Return the first episode that is not done
func getCurrentEpisode(userEpisodeCollection []api.UserEpisodeCollection) (api.UserEpisodeCollection, error) {
	doneType := api.EpisodeCollectionType["done"]
	for _, userEpisode := range userEpisodeCollection {
		if userEpisode.Type != doneType {
			return userEpisode, nil
		}
	}
	return userEpisodeCollection[0], fmt.Errorf("no more episodes to watch")
}
