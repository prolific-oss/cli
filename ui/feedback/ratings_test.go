package feedback_test

import (
	"testing"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui/feedback"
)

func TestNewRatingItems(t *testing.T) {
	average := 4.5
	ratings := client.StudyRatingsResponse{
		"clarity_rating": model.StudyRating{
			AverageRating: &average,
			TotalCount:    3,
		},
	}

	items := feedback.NewRatingItems(ratings)
	if len(items) != 3 {
		t.Fatalf("expected 3 rating rows, got %d", len(items))
	}

	if items[0].Rating != "clarity" || items[0].Average != "4.5" || items[0].Responses != 3 {
		t.Fatalf("unexpected clarity row: %#v", items[0])
	}

	if items[1].Rating != "ease" || items[1].Average != "-" || items[1].Responses != 0 {
		t.Fatalf("unexpected ease row: %#v", items[1])
	}

	if items[2].Rating != "fairness" {
		t.Fatalf("unexpected fairness row: %#v", items[2])
	}
}
