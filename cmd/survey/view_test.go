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

func TestNewViewCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := survey.NewViewCommand("view", c, os.Stdout)

	if cmd.Use != "view" {
		t.Fatalf("expected use: view; got %s", cmd.Use)
	}

	if cmd.Short != "Provide details about your survey" {
		t.Fatalf("expected short: Provide details about your survey; got %s", cmd.Short)
	}
}

func TestViewSurvey(t *testing.T) {
	tests := []struct {
		name             string
		surveyID         string
		mockReturn       *model.Survey
		mockError        error
		expectedContains []string
		expectedError    string
	}{
		{
			name:     "survey with questions",
			surveyID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			mockReturn: &model.Survey{
				ID:           "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				ResearcherID: testResearcherID,
				Title:        "Screening Survey",
				DateCreated:  time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
				DateModified: time.Date(2026, 1, 16, 14, 0, 0, 0, time.UTC),
				Questions: []model.SurveyQuestion{
					{
						ID:    "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
						Title: "What is your handedness?",
						Type:  "single",
						Answers: []model.SurveyAnswerOption{
							{ID: "f1e2d3c4-b5a6-7890-abcd-ef1234567891", Value: "Left"},
							{ID: "f1e2d3c4-b5a6-7890-abcd-ef1234567892", Value: "Right"},
							{ID: "f1e2d3c4-b5a6-7890-abcd-ef1234567893", Value: "Ambidextrous"},
						},
					},
				},
			},
			expectedContains: []string{"Screening Survey", "What is your handedness?", "Left", "Right", "Ambidextrous"},
		},
		{
			name:     "survey with sections",
			surveyID: "7ca8c921-0ebe-22e2-91c5-11d15ge541d9",
			mockReturn: &model.Survey{
				ID:           "7ca8c921-0ebe-22e2-91c5-11d15ge541d9",
				ResearcherID: testResearcherID,
				Title:        "Sectioned Survey",
				DateCreated:  time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				DateModified: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				Sections: []model.SurveySection{
					{
						ID:    "b2c3d4e5-f6a7-8901-bcde-f12345678901",
						Title: "Demographics",
						Questions: []model.SurveyQuestion{
							{
								ID:    "c3d4e5f6-a7b8-9012-cdef-123456789012",
								Title: "What is your age?",
								Type:  "single",
								Answers: []model.SurveyAnswerOption{
									{Value: "18-25"},
									{Value: "26-35"},
								},
							},
						},
					},
				},
			},
			expectedContains: []string{"Sectioned Survey", "Demographics", "What is your age?"},
		},
		{
			name:     "survey with no questions",
			surveyID: "afdb1254-3he1-55h5-c4f8-44g48jh874gc",
			mockReturn: &model.Survey{
				ID:           "afdb1254-3he1-55h5-c4f8-44g48jh874gc",
				ResearcherID: testResearcherID,
				Title:        "Empty Survey",
				DateCreated:  time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
				DateModified: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedContains: []string{"Empty Survey", "No questions defined"},
		},
		{
			name:          "API error",
			surveyID:      "8db9d032-1fcf-33f3-a2d6-22e26hf652ea",
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
				GetSurvey(gomock.Eq(tt.surveyID)).
				Return(tt.mockReturn, tt.mockError).
				Times(1)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := survey.NewViewCommand("view", c, writer)
			cmd.SetArgs([]string{tt.surveyID})
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
