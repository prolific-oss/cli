package feedback

import (
	"sort"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
)

// RatingFields are the fields shown in study ratings table and CSV output.
const RatingFields = "Rating,Average,Responses"

type studyRatingID struct {
	ID    string
	Label string
}

var studyRatingIDs = []studyRatingID{
	{ID: "clarity_rating", Label: "clarity"},
	{ID: "difficulty_rating", Label: "ease"},
	{ID: "fairness_rating", Label: "fairness"},
}

// RatingSummary is one aggregate study rating. Every output format renders from
// this so JSON, CSV and table agree on labels and on which ratings appear.
type RatingSummary struct {
	Rating    string   `json:"rating"`
	Average   *float64 `json:"average_rating"`
	Responses int      `json:"total_count"`
}

// RatingItem is the text-formatted view of a RatingSummary, used by the table
// and CSV renderers which read fields by name via reflection.
type RatingItem struct {
	Rating    string
	Average   string
	Responses int
}

// NewRatingSummaries flattens the ratings response into deterministic rows: the
// known ratings in display order, then any rating the API has added since, so a
// new key is never silently dropped from one output format.
func NewRatingSummaries(ratings client.StudyRatingsResponse) []RatingSummary {
	summaries := make([]RatingSummary, 0, len(ratings))
	seen := make(map[string]bool, len(studyRatingIDs))

	for _, rating := range studyRatingIDs {
		seen[rating.ID] = true
		summaries = append(summaries, newRatingSummary(rating.Label, ratings[rating.ID]))
	}

	extra := make([]string, 0)
	for id := range ratings {
		if !seen[id] {
			extra = append(extra, id)
		}
	}
	sort.Strings(extra)

	for _, id := range extra {
		summaries = append(summaries, newRatingSummary(id, ratings[id]))
	}

	return summaries
}

func newRatingSummary(label string, rating model.StudyRating) RatingSummary {
	return RatingSummary{
		Rating:    label,
		Average:   rating.AverageRating,
		Responses: rating.TotalCount,
	}
}

// NewRatingItems converts the ratings response into display-ready rows.
func NewRatingItems(ratings client.StudyRatingsResponse) []RatingItem {
	summaries := NewRatingSummaries(ratings)
	items := make([]RatingItem, 0, len(summaries))
	for _, summary := range summaries {
		items = append(items, RatingItem{
			Rating:    summary.Rating,
			Average:   formatRating(summary.Average),
			Responses: summary.Responses,
		})
	}
	return items
}
