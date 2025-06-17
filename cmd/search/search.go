package search

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/iucario/bangumi-go/api"
	"github.com/iucario/bangumi-go/cmd"
	"github.com/spf13/cobra"
)

var (
	keyword     string
	sort        string
	subjectType []string
	metaTags    []string
	tags        []string
	airDate     string
	rating      int
	rank        int
	nsfw        bool
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search bangumi",
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewAuthClientWithConfig()

		// Initialize filter
		filter := api.Filter{
			MetaTags: metaTags,
			Tag:      tags,
			AirDate:  airDate,
			Rating:   rating,
			Rank:     rank,
			NSFW:     nsfw,
		}

		// Convert string types to SubjectType array
		if len(subjectType) > 0 {
			types := make([]api.SubjectType, 0, len(subjectType))
			for _, t := range subjectType {
				if subjectT, ok := api.SubjectTypeMap[strings.ToLower(t)]; ok && subjectT != api.SubjectType(0) { // Exclude custom "all" type
					types = append(types, subjectT)
				} else {
					slog.Warn("Invalid subject type", "type", t)
				}
			}
			if len(types) > 0 {
				filter.Type = types
			}
		}

		// Convert sort string to Sort type
		var sortEnum, ok = api.SortMap[strings.ToLower(sort)]
		if !ok {
			slog.Error("Invalid sort option", "sort", sort)
			return
		}

		payload := api.Payload{
			Keyword: keyword,
			Sort:    sortEnum,
			Filter:  filter,
		}

		// Debug: print the request payload
		if reqBody, err := json.MarshalIndent(payload, "", "  "); err == nil {
			slog.Debug("Search request payload", "payload", string(reqBody))
		}
		pagesize := 20
		offset := 0
		result, err := Search(client, payload, pagesize, offset)
		if err != nil {
			slog.Error("Error searching", "error", err)
			return
		}

		// Print results in a formatted way
		fmt.Printf("Total results: %d\n", result.Total)
		fmt.Printf("Showing results %d/%d\n", result.Limit, result.Total)

		for _, subject := range result.Data {
			fmt.Printf("ID: %d Name: %s Type: %v Rating: %.1f (%d)\n",
				subject.ID, subject.GetName(), subject.Type, subject.Rating.Score, subject.Rating.Total)
		}
	},
}

func init() {
	searchCmd.Flags().StringVarP(&keyword, "keyword", "k", "", "Search keyword")
	searchCmd.Flags().StringVarP(&sort, "sort", "s", "match", "Sort by: match, heat, rank, score")
	searchCmd.Flags().StringSliceVarP(&subjectType, "type", "t", nil, "Subject types: book, anime, music, game, real (can specify multiple)")
	searchCmd.Flags().StringSliceVarP(&metaTags, "meta-tag", "m", nil, "Meta tags (AND relation). Add '-' at beginning to exclude")
	searchCmd.Flags().StringSliceVarP(&tags, "tag", "T", nil, "Tags (AND relation). Add '-' at beginning to exclude")
	searchCmd.Flags().StringVarP(&airDate, "air-date", "d", "", "Air date filter (YYYY-MM-DD)")
	searchCmd.Flags().IntVarP(&rating, "rating", "r", 0, "Rating filter")
	searchCmd.Flags().IntVarP(&rank, "rank", "R", 0, "Rank filter")
	searchCmd.Flags().BoolVarP(&nsfw, "nsfw", "n", false, "Include NSFW content")

	cmd.RootCmd.AddCommand(searchCmd)
}

func Search(c api.Client, payload api.Payload, limit, offset int) (*api.SubjectList, error) {
	url := "https://api.bgm.tv/v0/search/subjects"
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshalling payload: %v", err)
	}

	bytes, err := c.Post(url, reqBody)
	if err != nil {
		return nil, err
	}

	var result api.SubjectList
	if err := json.Unmarshal(bytes, &result); err != nil {
		return nil, fmt.Errorf("unmarshalling response: %v", err)
	}
	return &result, nil
}
