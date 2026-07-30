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

func TestNewRatingSummariesSharesLabelsWithItems(t *testing.T) {
	average := 4.5
	ratings := client.StudyRatingsResponse{
		"clarity_rating": model.StudyRating{AverageRating: &average, TotalCount: 3},
	}

	summaries := feedback.NewRatingSummaries(ratings)
	items := feedback.NewRatingItems(ratings)

	if len(summaries) != len(items) {
		t.Fatalf("expected summaries and items to cover the same ratings, got %d and %d", len(summaries), len(items))
	}

	for i, summary := range summaries {
		if summary.Rating != items[i].Rating {
			t.Fatalf("row %d label mismatch: %q vs %q", i, summary.Rating, items[i].Rating)
		}
		if summary.Responses != items[i].Responses {
			t.Fatalf("row %d response count mismatch: %d vs %d", i, summary.Responses, items[i].Responses)
		}
	}

	if summaries[0].Average == nil || *summaries[0].Average != average {
		t.Fatalf("expected clarity average %v, got %#v", average, summaries[0].Average)
	}

	if summaries[2].Average != nil {
		t.Fatalf("expected missing fairness average to stay nil, got %#v", summaries[2].Average)
	}
}

func TestNewRatingSummariesIncludesUnknownRatings(t *testing.T) {
	ratings := client.StudyRatingsResponse{
		"enjoyment_rating": model.StudyRating{TotalCount: 7},
	}

	summaries := feedback.NewRatingSummaries(ratings)
	if len(summaries) != 4 {
		t.Fatalf("expected the unknown rating to be appended, got %d rows", len(summaries))
	}

	if summaries[3].Rating != "enjoyment_rating" || summaries[3].Responses != 7 {
		t.Fatalf("unexpected trailing row: %#v", summaries[3])
	}

	items := feedback.NewRatingItems(ratings)
	if len(items) != len(summaries) {
		t.Fatalf("table/CSV rows dropped the unknown rating: %d vs %d", len(items), len(summaries))
	}
}
