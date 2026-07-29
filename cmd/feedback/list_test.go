package feedback_test

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/feedback"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := feedback.NewListCommand(client, os.Stdout)

	use := "list"
	short := "List participant feedback for a study, requires Study ID"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected use: %s; got %s", short, cmd.Short)
	}
}

func TestNewListCommandRequiresStudyID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := feedback.NewListCommand(c, writer)
	err := cmd.RunE(cmd, []string{})

	if err == nil {
		t.Fatal("expected an error when no study ID is provided")
	}

	expected := "please provide a study ID"
	if err.Error() != expected {
		t.Fatalf("expected error '%s'; got '%s'", expected, err.Error())
	}
}

func TestNewListCommandRejectsMalformedStudyID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := feedback.NewListCommand(c, writer)
	_ = cmd.Flags().Set("study", "not-a-valid-id")
	err := cmd.RunE(cmd, []string{})

	if err == nil {
		t.Fatal("expected an error for a malformed study ID")
	}

	expected := `invalid study ID "not-a-valid-id": must be a 24-character hex string`
	if err.Error() != expected {
		t.Fatalf("expected error '%s'; got '%s'", expected, err.Error())
	}
}

func TestNewListCommandRejectsNegativeLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := feedback.NewListCommand(c, os.Stdout)
	_ = cmd.Flags().Set("study", "63c123af913a974f87e8e7fc")
	_ = cmd.Flags().Set("limit", "-1")

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected an error for a negative limit")
	}

	expected := "limit must be greater than or equal to 0"
	if err.Error() != expected {
		t.Fatalf("expected error '%s'; got '%s'", expected, err.Error())
	}
}

func TestNewListCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "63c123af913a974f87e8e7fc"
	clarity := 4.5
	ease := 3.0
	fairness := 5.0
	category := "study-not-as-described"
	text := "The task was confusing."

	feedbackResponse := client.ListStudyFeedbackResponse{
		Results: []model.StudyFeedback{
			{
				ParticipantID: "919",
				Category:      &category,
				Text:          &text,
				Ratings: model.StudyFeedbackRatings{
					Clarity:  &clarity,
					Ease:     &ease,
					Fairness: &fairness,
				},
			},
		},
		JSONAPIMeta: &client.JSONAPIMeta{
			Meta: struct {
				Count int `json:"count"`
			}{
				Count: 1,
			},
		},
	}

	ratingsResponse := client.StudyRatingsResponse{
		"clarity_rating":    model.StudyRating{AverageRating: &clarity, TotalCount: 1},
		"difficulty_rating": model.StudyRating{AverageRating: &ease, TotalCount: 1},
		"fairness_rating":   model.StudyRating{AverageRating: &fairness, TotalCount: 1},
	}

	c.
		EXPECT().
		GetStudyRatings(gomock.Eq(studyID)).
		Return(&ratingsResponse, nil).
		AnyTimes()

	c.
		EXPECT().
		GetStudyFeedback(gomock.Eq(studyID), gomock.Eq(false), gomock.Eq(0), gomock.Eq(0)).
		Return(&feedbackResponse, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := feedback.NewListCommand(c, writer)
	_ = cmd.Flags().Set("study", studyID)
	err := cmd.RunE(cmd, []string{studyID})

	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	writer.Flush()

	actual := stripansi.Strip(b.String())

	if !strings.Contains(actual, "919") {
		t.Fatalf("expected output to contain participant id, got: %s", actual)
	}

	if !strings.Contains(actual, "The task was confusing.") {
		t.Fatalf("expected output to contain written feedback text, got: %s", actual)
	}

	if !strings.Contains(actual, "Showing 1 record of 1") {
		t.Fatalf("expected output to contain record counter, got: %s", actual)
	}
}
