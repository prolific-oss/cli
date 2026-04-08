package survey_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/survey"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewResponseViewCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := survey.NewResponseViewCommand("view", c, os.Stdout)

	if cmd.Use != "view <survey_id> <response_id>" {
		t.Fatalf("expected use: view <survey_id> <response_id>; got %s", cmd.Use)
	}

	if cmd.Short != "View a survey response" {
		t.Fatalf("expected short: View a survey response; got %s", cmd.Short)
	}
}

func TestViewSurveyResponse(t *testing.T) {
	tests := []struct {
		name             string
		surveyID         string
		responseID       string
		mockReturn       *model.SurveyResponse
		mockError        error
		expectedContains []string
		expectedError    string
	}{
		{
			name:       "response with flat questions",
			surveyID:   testSurveyID,
			responseID: "resp-001",
			mockReturn: &model.SurveyResponse{
				ID:            "resp-001",
				ParticipantID: "part-001",
				SubmissionID:  "sub-001",
				DateCreated:   time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
				DateModified:  time.Date(2026, 1, 15, 10, 35, 0, 0, time.UTC),
				Questions: []model.SurveyQuestionResponse{
					{
						QuestionID:    "q-001",
						QuestionTitle: "What is your handedness?",
						Answers: []model.SurveyResponseAnswer{
							{AnswerID: "a-001", Value: "Left"},
						},
					},
				},
			},
			expectedContains: []string{"resp-001", "part-001", "sub-001", "What is your handedness?", "Left"},
		},
		{
			name:       "response with sections",
			surveyID:   testSurveyID,
			responseID: "resp-002",
			mockReturn: &model.SurveyResponse{
				ID:            "resp-002",
				ParticipantID: "part-002",
				SubmissionID:  "sub-002",
				DateCreated:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				DateModified:  time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				Sections: []model.SurveyResponseSection{
					{
						SectionID: "sec-001",
						Questions: []model.SurveyQuestionResponse{
							{
								QuestionID:    "q-001",
								QuestionTitle: "What is your age?",
								Answers: []model.SurveyResponseAnswer{
									{AnswerID: "a-001", Value: "26-35"},
								},
							},
						},
					},
				},
			},
			expectedContains: []string{"resp-002", "sec-001", "What is your age?", "26-35"},
		},
		{
			name:       "response with no answers",
			surveyID:   testSurveyID,
			responseID: "resp-003",
			mockReturn: &model.SurveyResponse{
				ID:            "resp-003",
				ParticipantID: "part-003",
				SubmissionID:  "sub-003",
				DateCreated:   time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
				DateModified:  time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedContains: []string{"resp-003", "No answers recorded"},
		},
		{
			name:          "API error",
			surveyID:      testSurveyID,
			responseID:    "resp-404",
			mockError:     errors.New("not found"),
			expectedError: "error: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			c.EXPECT().
				GetSurveyResponse(gomock.Eq(tt.surveyID), gomock.Eq(tt.responseID)).
				Return(tt.mockReturn, tt.mockError).
				Times(1)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := survey.NewResponseViewCommand("view", c, writer)
			cmd.SetArgs([]string{tt.surveyID, tt.responseID})
			err := cmd.Execute()

			writer.Flush()

			if tt.expectedError != "" {
				if err == nil || err.Error() != tt.expectedError {
					t.Fatalf("expected error '%s'; got '%v'", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := b.String()
			for _, expected := range tt.expectedContains {
				if !bytes.Contains([]byte(output), []byte(expected)) {
					t.Errorf("expected output to contain '%s', got:\n%s", expected, output)
				}
			}
		})
	}
}
