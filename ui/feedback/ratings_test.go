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

func TestNewRatingRows(t *testing.T) {
	average := 4.5
	ratings := client.StudyRatingsResponse{
		"clarity_rating": model.StudyRating{
			AverageRating: &average,
			TotalCount:    3,
		},
		"new_rating": model.StudyRating{
			AverageRating: &average,
			TotalCount:    1,
		},
	}

	rows := feedback.NewRatingRows(ratings)
	if len(rows) != 3 {
		t.Fatalf("expected 3 rating rows, got %d", len(rows))
	}

	if rows[0].Rating != "clarity" || rows[0].Average == nil || *rows[0].Average != 4.5 || rows[0].Responses != 3 {
		t.Fatalf("unexpected clarity row: %#v", rows[0])
	}

	if rows[1].Rating != "ease" || rows[1].Average != nil || rows[1].Responses != 0 {
		t.Fatalf("unexpected ease row: %#v", rows[1])
	}

	if rows[2].Rating != "fairness" {
		t.Fatalf("unexpected fairness row: %#v", rows[2])
	}
}
