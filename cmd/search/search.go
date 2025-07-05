package search

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
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
	airDate     []string
	rating      []string
	rank        []string
	nsfw        bool
)

var searchCmd = &cobra.Command{
	Use:     "search",
	Short:   "Search bangumi",
	Example: `bgm search -k "" -s match -c anime -T "原创" -t "科幻" -d ">=2020-01-01,<2025-12-31" -r ">=8,<10"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Show help if no input is provided
		if strings.TrimSpace(keyword) == "" && sort == "" && len(subjectType) == 0 && len(metaTags) == 0 &&
			len(tags) == 0 && len(airDate) == 0 && len(rating) == 0 && len(rank) == 0 {
			err := cmd.Help()
			if err != nil {
				fmt.Println("Unexpected error")
			}
			return
		}

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
		sortEnum, ok := api.SortMap[strings.ToLower(sort)]
		if !ok {
			slog.Error("Invalid sort option", "sort", sort)
			sortEnum = api.MATCH // Default to match if invalid
		}

		payload := api.Payload{
			Keyword: keyword,
			Sort:    sortEnum,
			Filter:  filter,
		}

		// Debug: print the request payload
		slog.Debug("Search request payload", "payload", payload)

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
			fmt.Printf("%d %s | %v | %.1f (%d) | #%d\n",
				subject.ID, subject.GetName(), api.SubjectTypeRev[int(subject.Type)], subject.Rating.Score, subject.Rating.Total, subject.Rating.Rank)
		}
	},
}

func init() {
	searchCmd.Flags().StringVarP(&keyword, "keyword", "k", "", "Search keyword")
	searchCmd.Flags().StringVarP(&sort, "sort", "s", "", "Sort by: match, heat, rank, score")
	searchCmd.Flags().StringSliceVarP(&subjectType, "type", "c", nil, "Subject types: book, anime, music, game, real (can select multiple) E.g. -c 'anime,book'")
	searchCmd.Flags().StringSliceVarP(&metaTags, "meta-tag", "T", nil, "Meta tags (AND relation). Add '-' at beginning to exclude E.g. -T '原创,热血,-OVA'")
	searchCmd.Flags().StringSliceVarP(&tags, "tag", "t", nil, "Tags (AND relation). Add '-' at beginning to exclude E.g. -t '原创,京阿尼,-OVA'")
	searchCmd.Flags().StringSliceVarP(&airDate, "date", "d", nil, "Air date filter E.g. -d '>=2010-01-01,<2010-12-31'")
	searchCmd.Flags().StringSliceVarP(&rating, "rating", "r", nil, "Rating filter E.g. -r '>=7,<9'")
	searchCmd.Flags().StringSliceVarP(&rank, "rank", "R", nil, "Rank filter E.g. -R '<=200,>100'")
	searchCmd.Flags().BoolVarP(&nsfw, "nsfw", "n", false, "Include NSFW content")

	cmd.RootCmd.AddCommand(searchCmd)
}

func Search(c api.Client, payload api.Payload, limit, offset int) (*api.SubjectList, error) {
	// Validate rating filter using regex
	ratingRegex := regexp.MustCompile(`^(>|<|>=|<=|=) *\d+(\.\d+)?$`)
	for _, r := range payload.Filter.Rating {
		if !ratingRegex.MatchString(r) {
			return nil, fmt.Errorf("invalid score filter: %q, must have \"(>|<|>=|<=|=)\" in the beginning of values", r)
		}
	}

	url := fmt.Sprintf("https://api.bgm.tv/v0/search/subjects?limit=%d&offset=%d", limit, offset)
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
