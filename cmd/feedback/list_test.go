package feedback_test

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/acarl005/stripansi"
	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/feedback"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

const feedbackTestStudyID = "study-id"

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

func TestNewListCommandPassesStudyIDToAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "not-a-valid-id"
	feedbackResponse := client.ListStudyFeedbackResponse{
		Results: []model.StudyFeedback{},
		JSONAPIMeta: &client.JSONAPIMeta{
			Meta: struct {
				Count int `json:"count"`
			}{},
		},
	}

	c.EXPECT().
		GetStudyFeedback(
			gomock.Eq(studyID),
			gomock.Eq(false),
			gomock.Eq(client.DefaultRecordLimit),
			gomock.Eq(client.DefaultRecordOffset),
		).
		Return(&feedbackResponse, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := feedback.NewListCommand(c, writer)
	_ = cmd.Flags().Set("study", studyID)
	_ = cmd.Flags().Set("table", "true")
	err := cmd.RunE(cmd, []string{})

	if err != nil {
		t.Fatalf("did not expect error, got %v", err)
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

	c.
		EXPECT().
		GetStudyFeedback(
			gomock.Eq(studyID),
			gomock.Eq(false),
			gomock.Eq(client.DefaultRecordLimit),
			gomock.Eq(client.DefaultRecordOffset),
		).
		Return(&feedbackResponse, nil).
		AnyTimes()

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := feedback.NewListCommand(c, writer)
	_ = cmd.Flags().Set("study", studyID)
	_ = cmd.Flags().Set("table", "true")
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

func TestNewListCommandSupportsStandardOutputFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	cmd := feedback.NewListCommand(c, writer)

	for shorthand, name := range map[string]string{
		"j": "json",
		"t": "table",
		"c": "csv",
		"n": "non-interactive",
	} {
		flag := cmd.Flags().ShorthandLookup(shorthand)
		if flag == nil || flag.Name != name {
			t.Fatalf("expected -%s to map to --%s", shorthand, name)
		}
	}

	nonInteractive := cmd.Flags().Lookup("non-interactive")
	if nonInteractive == nil || !nonInteractive.Hidden {
		t.Fatal("expected --non-interactive to be a hidden compatibility flag")
	}
}

func TestNewListCommandMachineReadableOutput(t *testing.T) {
	tests := []struct {
		name     string
		flag     string
		response client.ListStudyFeedbackResponse
		expected []string
	}{
		{
			name: "JSON preserves nested ratings",
			flag: "json",
			response: func() client.ListStudyFeedbackResponse {
				clarity := 4.5
				return client.ListStudyFeedbackResponse{
					Results: []model.StudyFeedback{
						{
							ParticipantID: "919",
							Ratings: model.StudyFeedbackRatings{
								Clarity: &clarity,
							},
						},
					},
				}
			}(),
			expected: []string{`"participant_id": "919"`, `"ratings": {`, `"clarity": 4.5`},
		},
		{
			name: "CSV uses flat presentation rows",
			flag: "csv",
			response: func() client.ListStudyFeedbackResponse {
				clarity := 4.5
				return client.ListStudyFeedbackResponse{
					Results: []model.StudyFeedback{
						{
							ParticipantID: "919",
							Ratings: model.StudyFeedbackRatings{
								Clarity: &clarity,
							},
						},
					},
				}
			}(),
			expected: []string{
				"ParticipantID,Category,Clarity,Ease,Fairness,Text",
				"919,-,4.5,-,-,-",
			},
		},
		{
			name:     "empty JSON is an array",
			flag:     "json",
			response: client.ListStudyFeedbackResponse{},
			expected: []string{"[]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			c.EXPECT().
				GetStudyFeedback(
					gomock.Eq(feedbackTestStudyID),
					gomock.Eq(false),
					gomock.Eq(client.DefaultRecordLimit),
					gomock.Eq(client.DefaultRecordOffset),
				).
				Return(&tt.response, nil)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)
			cmd := feedback.NewListCommand(c, writer)
			_ = cmd.Flags().Set("study", feedbackTestStudyID)
			_ = cmd.Flags().Set(tt.flag, "true")

			if err := cmd.RunE(cmd, nil); err != nil {
				t.Fatalf("did not expect error, got %v", err)
			}
			writer.Flush()

			actual := b.String()
			for _, expected := range tt.expected {
				if !strings.Contains(actual, expected) {
					t.Fatalf("expected output to contain %q, got: %s", expected, actual)
				}
			}
		})
	}
}

func TestNewListCommandReturnsAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.EXPECT().
		GetStudyFeedback(
			gomock.Eq(feedbackTestStudyID),
			gomock.Eq(false),
			gomock.Eq(client.DefaultRecordLimit),
			gomock.Eq(client.DefaultRecordOffset),
		).
		Return(nil, errors.New("API error"))

	cmd := feedback.NewListCommand(c, &bytes.Buffer{})
	_ = cmd.Flags().Set("study", feedbackTestStudyID)
	_ = cmd.Flags().Set("json", "true")

	err := cmd.RunE(cmd, nil)
	if err == nil || err.Error() != "error: API error" {
		t.Fatalf("expected API error, got %v", err)
	}
}
