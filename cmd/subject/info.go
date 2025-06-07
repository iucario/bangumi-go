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
		subject, err := GetSubjectInfo(authClient, subjectId)
		if err != nil {
			fmt.Println("error", err)
		} else {
			fmt.Printf("%d\n%s\n%s\n%s\n", subject.ID, subject.NameCn, subject.Name, subject.Summary)
		}
	},
}

func init() {
	subCmd.AddCommand(infoCmd)
}

func GetSubjectInfo(c api.Client, subjectId int) (*api.Subject, error) {
	url := fmt.Sprintf("https://api.bgm.tv/v0/subjects/%d", subjectId)

	b, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	subject := api.Subject{}
	err = json.Unmarshal(b, &subject)
	if err != nil {
		return nil, err
	}
	return &subject, nil
}
