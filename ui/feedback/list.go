package feedback

import (
	"fmt"

	"github.com/prolific-oss/cli/model"
)

// ListFields are the feedback fields shown in table and CSV output.
const ListFields = "ParticipantID,Category,Clarity,Difficulty,Fairness,Text"

// ListItem is the flat presentation model used by table and CSV renderers.
// The API model remains nested so JSON output preserves the API response shape.
type ListItem struct {
	ParticipantID string
	Category      string
	Clarity       string
	Difficulty    string
	Fairness      string
	Text          string
}

// NewListItems converts feedback API models into display-ready rows.
func NewListItems(records []model.StudyFeedback) []ListItem {
	items := make([]ListItem, 0, len(records))
	for _, record := range records {
		items = append(items, ListItem{
			ParticipantID: formatOptionalString(record.ParticipantID),
			Category:      formatOptionalString(record.Category),
			Clarity:       formatRating(record.Ratings.Clarity),
			Difficulty:    formatRating(record.Ratings.Difficulty),
			Fairness:      formatRating(record.Ratings.Fairness),
			Text:          formatOptionalString(record.Text),
		})
	}
	return items
}

func formatRating(rating *float64) string {
	if rating == nil {
		return "-"
	}
	return fmt.Sprintf("%v", *rating)
}

func formatOptionalString(value *string) string {
	if value == nil || *value == "" {
		return "-"
	}
	return *value
}
