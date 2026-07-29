package survey_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/survey"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewResponseSummaryCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := survey.NewResponseSummaryCommand("summary", c, os.Stdout)

	if cmd.Use != "summary <survey_id>" {
		t.Fatalf("expected use: summary <survey_id>; got %s", cmd.Use)
	}

	if cmd.Short != "View a summary of survey responses" {
		t.Fatalf("expected short: View a summary of survey responses; got %s", cmd.Short)
	}
}

func TestViewSurveyResponseSummary(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		mockReturn       *model.SurveySummary
		mockError        error
		expectedContains []string
		expectedError    string
	}{
		{
			name: "summary with questions",
			args: []string{testSurveyID},
			mockReturn: &model.SurveySummary{
				SurveyID: testSurveyID,
				Questions: []model.SurveySummaryQuestion{
					{
						QuestionID:   "q-001",
						Question:     "What is your handedness?",
						TotalAnswers: 100,
						Answers: []model.SurveySummaryAnswer{
							{AnswerID: "a-001", Answer: "Left", Count: 15},
							{AnswerID: "a-002", Answer: "Right", Count: 80},
							{AnswerID: "a-003", Answer: "Ambidextrous", Count: 5},
						},
					},
				},
			},
			expectedContains: []string{
				"Survey Response Summary",
				testSurveyID,
				"What is your handedness?",
				"Total Answers: 100",
				"Left: 15",
				"Right: 80",
				"Ambidextrous: 5",
			},
		},
		{
			name: "summary with no responses",
			args: []string{testSurveyID},
			mockReturn: &model.SurveySummary{
				SurveyID:  testSurveyID,
				Questions: []model.SurveySummaryQuestion{},
			},
			expectedContains: []string{"No responses recorded"},
		},
		{
			name: "json output",
			args: []string{testSurveyID, "--json"},
			mockReturn: &model.SurveySummary{
				SurveyID: testSurveyID,
				Questions: []model.SurveySummaryQuestion{
					{
						QuestionID:   "q-001",
						Question:     "Handedness?",
						TotalAnswers: 10,
						Answers: []model.SurveySummaryAnswer{
							{AnswerID: "a-001", Answer: "Left", Count: 3},
						},
					},
				},
			},
			expectedContains: []string{testSurveyID, "q-001", "Handedness?", "total_answers", "10"},
		},
		{
			name:          "API error",
			args:          []string{testSurveyID},
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
				GetSurveyResponseSummary(gomock.Eq(testSurveyID)).
				Return(tt.mockReturn, tt.mockError).
				Times(1)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := survey.NewResponseSummaryCommand("summary", c, writer)
			cmd.SetArgs(tt.args)
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
