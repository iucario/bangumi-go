package subject

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/iucario/bangumi-go/api"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Subject information",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		subjectId, err := strconv.Atoi(id)
		if err != nil {
			slog.Error(fmt.Sprintf("Invalid subject ID: %s", id))
			return
		}
		authClient := api.NewAuthClientWithConfig()
		subject := GetSubjectInfo(authClient, subjectId)
		fmt.Printf("%d\n%s\n%s\n%s\n", subject.ID, subject.NameCn, subject.Name, subject.Summary)
	},
}

func init() {
	subCmd.AddCommand(infoCmd)
}

func GetSubjectInfo(c api.Client, subjectId int) *api.Subject {
	url := fmt.Sprintf("https://api.bgm.tv/v0/subjects/%d", subjectId)

	b, err := c.Get(url)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get subject info: %v", err))
		return nil
	}
	subject := api.Subject{}
	err = json.Unmarshal(b, &subject)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to unmarshal subject info: %v", err))
		return nil
	}
	return &subject
}
