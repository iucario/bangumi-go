package subject

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/iucario/bangumi-go/api"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Subject collection status",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		subjectId, err := strconv.Atoi(id)
		if err != nil {
			slog.Error(fmt.Sprintf("Invalid subject ID: %s", id))
			return
		}

		status, _ := cmd.Flags().GetString("status")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		rate, _ := cmd.Flags().GetInt("rate")
		comment, _ := cmd.Flags().GetString("comment")
		private, _ := cmd.Flags().GetBool("private")

		credential, _ := api.LoadCredential(ConfigDir)
		userInfo, err := api.GetUserInfo(credential.AccessToken)
		api.AbortOnError(err)
		collection, _ := GetUserSubjectCollection(credential.AccessToken, userInfo.Username, subjectId)

		modifyCollection(credential.AccessToken, subjectId, status, tags, rate, comment, private, collection)
		subject := GetSubjectInfo(subjectId)
		fmt.Printf("%d\n%s\n%s\n", subject.ID, subject.NameCn, subject.Name)

		printSubjectStatus(credential.AccessToken, subjectId, collection)
	},
}

func init() {
	var status string
	var tags []string
	var rate int
	var comment string
	var private bool
	statusCmd.Flags().StringVarP(&status, "status", "s", "", "Status: wish, done, watch, onhold, dropped")
	statusCmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "Tags")
	statusCmd.Flags().IntVarP(&rate, "rate", "r", 0, "Rating")
	statusCmd.Flags().StringVarP(&comment, "comment", "c", "", "Comment")
	statusCmd.Flags().BoolVarP(&private, "private", "p", false, "Private")
	subCmd.AddCommand(statusCmd)
}

// Modify collection if any of the args are not empty and different from the current collection
func modifyCollection(token string, subjectId int, status string, tags []string, rate int, comment string, private bool, collection api.UserSubjectCollection) {
	slog.Info(fmt.Sprintf("called modifyCollection: %s %v %d %s private %v", status, tags, rate, comment, private))
	finalStatus := api.CollectionTypeRev[int(collection.Type)]
	if status != "" && validateStatus(status) {
		finalStatus = status
	}
	finalTags := collection.Tags
	if len(tags) > 0 {
		finalTags = tags
	}
	finalRate := int(collection.Rate)
	if rate > 0 {
		finalRate = rate
	}
	finalComment := collection.Comment
	if comment != "" {
		finalComment = comment
	}
	finalPrivate := private

	if collectionChanged(finalStatus, finalTags, finalRate, finalComment, finalPrivate, collection) {
		err := PostCollection(token, subjectId, finalStatus, finalTags, finalComment, finalRate, finalPrivate)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to modify collection: %v", err))
		} else {
			slog.Info(fmt.Sprintf("Successfully modified collection for subject %d", subjectId))
		}
	}
}

func collectionChanged(status string, tags []string, rate int, comment string, private bool, collection api.UserSubjectCollection) bool {
	if status != api.CollectionTypeRev[int(collection.Type)] {
		return true
	}
	if !sameStringSlices(tags, collection.Tags) {
		return true
	}
	if rate != int(collection.Rate) {
		return true
	}
	if comment != collection.Comment {
		return true
	}
	if private != collection.Private {
		return true
	}
	return false
}

func validateStatus(status string) bool {
	return status == "wish" || status == "done" || status == "watch" || status == "onhold" || status == "dropped"
}

func printSubjectStatus(token string, subjectId int, collection api.UserSubjectCollection) {
	tags := strings.Join(collection.Tags, ", ")
	fmt.Printf("Status: %s\n", api.CollectionTypeRev[int(collection.Type)])
	fmt.Printf("Subjcet type: %s\n", api.SubjectTypeRev[int(collection.Subject.Type)])
	fmt.Printf("Your Tags: %s\n", tags)
	fmt.Printf("Your Rating: %d\n", collection.Rate)

	userEpisodes, _ := GetUserEpisodeCollections(token, subjectId, 0, 100, 0)
	status := getEpisodeStatus(&userEpisodes.Data)
	printEpisodeStatus(status)
}

func getEpisodeStatus(userEpisodeCollection *[]api.UserEpisodeCollection) []int {
	status := make([]int, len(*userEpisodeCollection))
	for i, userEpisode := range *userEpisodeCollection {
		status[i] = int(userEpisode.Type)
	}
	return status
}

func printEpisodeStatus(status []int) {
	for i, s := range status {
		epNum := fmt.Sprintf("%02d", i+1)
		if s == 2 {
			fmt.Printf("\033[47;30m%s\033[0m", epNum) // White background, black text
			fmt.Print(" ")
		} else {
			fmt.Printf("%s ", epNum)
		}
	}
	fmt.Println()
}

func sameStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
