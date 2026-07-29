package feedback_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
	"github.com/prolific-oss/cli/ui/feedback"
)

func sampleFeedbackResponse() client.ListStudyFeedbackResponse {
	clarity := 4.0
	ease := 3.0
	fairness := 5.0
	category := "study-not-as-described"
	text := "Instructions, were confusing."

	return client.ListStudyFeedbackResponse{
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
}

func sampleRatingsResponse() client.StudyRatingsResponse {
	average := 4.0
	return client.StudyRatingsResponse{
		"clarity_rating":    model.StudyRating{AverageRating: &average, TotalCount: 3},
		"difficulty_rating": model.StudyRating{AverageRating: &average, TotalCount: 3},
		"fairness_rating":   model.StudyRating{AverageRating: &average, TotalCount: 3},
	}
}

func TestNonInteractiveRendererRendersRatingsAndFeedback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	opts := feedback.ListUsedOptions{
		StudyID: "1234",
		Limit:   0,
		Offset:  0,
	}

	feedbackResponse := sampleFeedbackResponse()
	ratingsResponse := sampleRatingsResponse()

	c.EXPECT().GetStudyRatings(gomock.Eq(opts.StudyID)).Return(&ratingsResponse, nil).MaxTimes(1)
	c.EXPECT().
		GetStudyFeedback(gomock.Eq(opts.StudyID), gomock.Eq(false), gomock.Eq(opts.Limit), gomock.Eq(opts.Offset)).
		Return(&feedbackResponse, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	renderer := feedback.NonInteractiveRenderer{}
	err := renderer.Render(c, opts, writer)
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	writer.Flush()

	actual := b.String()

	if !strings.Contains(actual, "919") {
		t.Fatalf("expected participant id in output, got: %s", actual)
	}

	if !strings.Contains(actual, "Instructions, were confusing.") {
		t.Fatalf("expected written feedback text in output, got: %s", actual)
	}

	if !strings.Contains(actual, "Showing 1 record of 1") {
		t.Fatalf("expected record counter in output, got: %s", actual)
	}
}

func TestNonInteractiveRendererReturnsErrorForEmptyFeedback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	opts := feedback.ListUsedOptions{
		StudyID: "1234",
		Limit:   0,
		Offset:  0,
	}

	ratingsResponse := client.StudyRatingsResponse{}
	emptyResponse := client.ListStudyFeedbackResponse{
		Results: []model.StudyFeedback{},
		JSONAPIMeta: &client.JSONAPIMeta{
			Meta: struct {
				Count int `json:"count"`
			}{
				Count: 0,
			},
		},
	}

	c.EXPECT().GetStudyRatings(gomock.Eq(opts.StudyID)).Return(&ratingsResponse, nil).MaxTimes(1)
	c.EXPECT().
		GetStudyFeedback(gomock.Eq(opts.StudyID), gomock.Eq(false), gomock.Eq(opts.Limit), gomock.Eq(opts.Offset)).
		Return(&emptyResponse, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	renderer := feedback.NonInteractiveRenderer{}
	err := renderer.Render(c, opts, writer)
	if err == nil {
		t.Fatal("expected an error for empty feedback")
	}

	expected := "no feedback found for this study"
	if err.Error() != expected {
		t.Fatalf("expected error %q, got %q", expected, err.Error())
	}
}

func TestJSONRendererReturnsErrorForEmptyFeedback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	opts := feedback.ListUsedOptions{StudyID: "1234"}
	emptyResponse := client.ListStudyFeedbackResponse{
		Results: []model.StudyFeedback{},
		JSONAPIMeta: &client.JSONAPIMeta{Meta: struct {
			Count int `json:"count"`
		}{Count: 0}},
	}

	c.EXPECT().GetStudyFeedback(opts.StudyID, false, 0, 0).Return(&emptyResponse, nil)

	var b bytes.Buffer
	renderer := feedback.JSONRenderer{}
	err := renderer.Render(c, opts, &b)
	if err == nil {
		t.Fatal("expected an error for empty feedback")
	}
	if b.Len() != 0 {
		t.Fatalf("expected no machine-readable output on error, got %q", b.String())
	}
}

func TestCsvRendererRendersInCsvFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	opts := feedback.ListUsedOptions{
		StudyID: "1234",
		Limit:   0,
		Offset:  0,
	}

	feedbackResponse := sampleFeedbackResponse()

	c.EXPECT().
		GetStudyFeedback(gomock.Eq(opts.StudyID), gomock.Eq(false), gomock.Eq(opts.Limit), gomock.Eq(opts.Offset)).
		Return(&feedbackResponse, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	renderer := feedback.CsvRenderer{}
	err := renderer.Render(c, opts, writer)
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	writer.Flush()

	expected := "ParticipantID,Category,Clarity,Ease,Fairness,Text\n" +
		"919,study-not-as-described,4,3,5,\"Instructions, were confusing.\"\n"

	if b.String() != expected {
		t.Fatalf("expected '%v', got '%v'", expected, b.String())
	}
}

func TestJSONRendererRendersInJSONFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	opts := feedback.ListUsedOptions{
		StudyID: "1234",
		Limit:   0,
		Offset:  0,
	}

	feedbackResponse := sampleFeedbackResponse()

	c.EXPECT().
		GetStudyFeedback(gomock.Eq(opts.StudyID), gomock.Eq(false), gomock.Eq(opts.Limit), gomock.Eq(opts.Offset)).
		Return(&feedbackResponse, nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	renderer := feedback.JSONRenderer{}
	err := renderer.Render(c, opts, writer)
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	writer.Flush()

	var result []model.StudyFeedback
	if err := json.Unmarshal(b.Bytes(), &result); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 feedback record, got %d", len(result))
	}

	if result[0].ParticipantID != "919" {
		t.Fatalf("expected participant_id 919, got %s", result[0].ParticipantID)
	}

	if result[0].Ratings.Clarity == nil || *result[0].Ratings.Clarity != 4.0 {
		t.Fatalf("expected ratings.clarity 4.0, got %v", result[0].Ratings.Clarity)
	}
}

func TestFetchFeedbackOmitsLimitAndOffsetByDefault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	// Zero-value Limit/Offset - the caller hasn't asked to page through
	// results, so every record for the study should come back in one call.
	opts := feedback.ListUsedOptions{StudyID: "1234"}

	category := "other"
	text1 := "first"
	text2 := "second"

	everything := client.ListStudyFeedbackResponse{
		Results: []model.StudyFeedback{
			{ParticipantID: "1", Category: &category, Text: &text1},
			{ParticipantID: "2", Category: &category, Text: &text2},
		},
		JSONAPIMeta: &client.JSONAPIMeta{Meta: struct {
			Count int `json:"count"`
		}{Count: 2}},
	}

	c.EXPECT().GetStudyFeedback(opts.StudyID, false, 0, 0).Return(&everything, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	renderer := feedback.JSONRenderer{}
	err := renderer.Render(c, opts, writer)
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}

	writer.Flush()

	var result []model.StudyFeedback
	if err := json.Unmarshal(b.Bytes(), &result); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 feedback records, got %d", len(result))
	}
}

func TestFetchFeedbackForwardsExplicitLimitAndOffset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	opts := feedback.ListUsedOptions{StudyID: "1234", Limit: 10, Offset: 5}

	feedbackResponse := sampleFeedbackResponse()

	c.EXPECT().GetStudyFeedback(opts.StudyID, false, 10, 5).Return(&feedbackResponse, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	renderer := feedback.JSONRenderer{}
	err := renderer.Render(c, opts, writer)
	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
	}
}
