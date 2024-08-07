package subject

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd/auth"
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
		credential, _ := auth.LoadCredential()
		userInfo, err := auth.GetUserInfo(credential.AccessToken)
		auth.Check(err)
		subject := GetSubjectInfo(subjectId)
		fmt.Printf("%d\n%s\n%s\n", subject.ID, subject.NameCn, subject.Name)

		printSubjectStatus(credential.AccessToken, userInfo.Username, subjectId)
	},
}

func init() {
	subCmd.AddCommand(statusCmd)
}

func printSubjectStatus(token, username string, subjectId int) {
	collection, _ := GetUserSubjectCollection(token, username, subjectId)
	tags := strings.Join(collection.Tags, ", ")
	fmt.Printf("Status: %s\n", api.CollectionTypeRev[int(collection.Type)])
	fmt.Printf("Subjcet type: %s\n", api.SubjectTypeRev[int(collection.Subject.Type)])
	fmt.Printf("Your Tags: %s\n", tags)
	fmt.Printf("Your Rating: %d\n", collection.Rate)

	userEpisodes, _ := GetUserEpisodeCollections(token, username, subjectId, 0, 100, 0)
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
