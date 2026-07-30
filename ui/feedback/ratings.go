package feedback

import "github.com/prolific-oss/cli/client"

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

// RatingRow is the presentation model for one aggregate study rating in JSON
// output. Average is kept as *float64 so JSON renders null/a number instead
// of the "-" placeholder the table/CSV formats use.
type RatingRow struct {
	Rating    string   `json:"rating"`
	Average   *float64 `json:"average"`
	Responses int      `json:"responses"`
}

// NewRatingRows converts the ratings response map into deterministic rows
// for JSON output, using the same rating IDs/labels as NewRatingItems.
func NewRatingRows(ratings client.StudyRatingsResponse) []RatingRow {
	rows := make([]RatingRow, 0, len(studyRatingIDs))
	for _, rating := range studyRatingIDs {
		summary := ratings[rating.ID]
		rows = append(rows, RatingRow{
			Rating:    rating.Label,
			Average:   summary.AverageRating,
			Responses: summary.TotalCount,
		})
	}
	return rows
}

// RatingItem is the presentation model for one aggregate study rating.
type RatingItem struct {
	Rating    string
	Average   string
	Responses int
}

// NewRatingItems converts the ratings response map into deterministic rows.
func NewRatingItems(ratings client.StudyRatingsResponse) []RatingItem {
	items := make([]RatingItem, 0, len(studyRatingIDs))
	for _, rating := range studyRatingIDs {
		summary := ratings[rating.ID]
		items = append(items, RatingItem{
			Rating:    rating.Label,
			Average:   formatRating(summary.AverageRating),
			Responses: summary.TotalCount,
		})
	}
	return items
}
