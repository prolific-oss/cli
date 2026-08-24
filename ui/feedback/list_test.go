package feedback_test

import (
	"strings"
	"testing"

	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui/feedback"
)

func sampleFeedbackResponse() client.ListStudyFeedbackResponse {
	clarity := 4.0
	difficulty := 3.0
	fairness := 5.0
	category := "study-not-as-described"
	text := "Instructions, were confusing."
	participantID := "919"

	return client.ListStudyFeedbackResponse{
		Results: []model.StudyFeedback{
			{
				ParticipantID: &participantID,
				Category:      &category,
				Text:          &text,
				Ratings: model.StudyFeedbackRatings{
					Clarity:    &clarity,
					Difficulty: &difficulty,
					Fairness:   &fairness,
				},
			},
			{},
		},
	}
}

func TestNewListItems(t *testing.T) {
	items := feedback.NewListItems(sampleFeedbackResponse().Results)

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	first := items[0]
	if first.ParticipantID != "919" ||
		first.Category != "study-not-as-described" ||
		first.Clarity != "4" ||
		first.Difficulty != "3" ||
		first.Fairness != "5" ||
		first.Text != "Instructions, were confusing." {
		t.Fatalf("unexpected first item: %#v", first)
	}

	second := items[1]
	if second.ParticipantID != "-" ||
		second.Category != "-" ||
		second.Clarity != "-" ||
		second.Difficulty != "-" ||
		second.Fairness != "-" ||
		second.Text != "-" {
		t.Fatalf("expected missing values to render as dashes, got %#v", second)
	}
}

func TestListViewRendersMissingParticipantID(t *testing.T) {
	record := sampleFeedbackResponse().Results[1]

	view := feedback.ListView{Feedback: &record}

	if actual := view.View(); !strings.Contains(actual, "Participant: -") {
		t.Fatalf("expected missing participant ID to render as a dash, got: %s", actual)
	}
}

func TestListViewRendersSelectedFeedback(t *testing.T) {
	record := sampleFeedbackResponse().Results[0]

	view := feedback.ListView{
		Feedback: &record,
	}

	actual := view.View()
	for _, expected := range []string{
		"Participant feedback",
		"Participant: 919",
		"Clarity:     4",
		"Text:        Instructions, were confusing.",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected output to contain %q, got: %s", expected, actual)
		}
	}

	if strings.Contains(actual, "Study ratings") {
		t.Fatalf("did not expect aggregate ratings in feedback list detail, got: %s", actual)
	}
}
