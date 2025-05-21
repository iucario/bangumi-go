package calendar

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd"
	"github.com/spf13/cobra"
)

var calendarCmd = &cobra.Command{
	Use:   "cal",
	Short: "Show calendar",
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewHTTPClient("")
		calendars, err := GetCalendar(client)
		if err != nil {
			cmd.PrintErr(err)
			return
		}

		for _, cal := range calendars {
			// Sort items by follower count
			sort.Slice(cal.Items, func(i, j int) bool {
				followersI := cal.Items[i].CollectionCount.Wish + cal.Items[i].CollectionCount.Watching + cal.Items[i].CollectionCount.Done
				followersJ := cal.Items[j].CollectionCount.Wish + cal.Items[j].CollectionCount.Watching + cal.Items[j].CollectionCount.Done
				return followersI > followersJ
			})

			// Print header
			weekdayTitle := fmt.Sprintf("%d %s", cal.Weekday.ID, cal.Weekday.EN)
			fmt.Printf("%13s\033[1;36m%s\033[0m\n", "", weekdayTitle)
			fmt.Println(strings.Repeat("─", 30)) // Fixed width divider

			// Print items with right-aligned numbers
			for _, item := range cal.Items {
				name := item.Name
				if item.NameCn != "" {
					name = item.NameCn
				}
				// FIXME: API error, there is only Watching count
				followers := item.CollectionCount.Wish + item.CollectionCount.Watching + item.CollectionCount.Done
				fmt.Printf("%6d + %s\n", followers, name)
			}
			fmt.Println()
		}
	},
}

func GetCalendar(client *api.HTTPClient) ([]api.Calendar, error) {
	url := "https://api.bgm.tv/calendar" // Calendar API endpoint has no '/v0'
	bytes, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	var calendars []api.Calendar
	if err := json.Unmarshal(bytes, &calendars); err != nil {
		return nil, err
	}
	return calendars, nil
}

func init() {
	cmd.RootCmd.AddCommand(calendarCmd)
}
