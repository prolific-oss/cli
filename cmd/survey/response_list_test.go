package survey_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/survey"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewResponseListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := survey.NewResponseListCommand("list", c, os.Stdout)

	if cmd.Use != "list <survey_id>" {
		t.Fatalf("expected use: list <survey_id>; got %s", cmd.Use)
	}

	if cmd.Short != "List responses for a survey" {
		t.Fatalf("expected short: List responses for a survey; got %s", cmd.Short)
	}
}

func TestListSurveyResponses(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		responsesReturn  *client.ListSurveyResponsesResponse
		responsesError   error
		expectedOutput   string
		expectedContains []string
		expectedError    string
	}{
		{
			name: "table output",
			args: []string{testSurveyID, "--table"},
			responsesReturn: &client.ListSurveyResponsesResponse{
				Results: []model.SurveyResponse{
					{
						ID:            "resp-001",
						ParticipantID: "part-001",
						SubmissionID:  "sub-001",
						DateCreated:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:            "resp-002",
						ParticipantID: "part-002",
						SubmissionID:  "sub-002",
						DateCreated:   time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			expectedContains: []string{"resp-001", "part-001", "sub-001", "resp-002", "part-002", "sub-002"},
		},
		{
			name: "json output",
			args: []string{testSurveyID, "--json"},
			responsesReturn: &client.ListSurveyResponsesResponse{
				Results: []model.SurveyResponse{
					{
						ID:            "resp-001",
						ParticipantID: "part-001",
						SubmissionID:  "sub-001",
					},
				},
			},
			expectedContains: []string{"resp-001", "part-001", "sub-001"},
		},
		{
			name: "csv output",
			args: []string{testSurveyID, "--csv"},
			responsesReturn: &client.ListSurveyResponsesResponse{
				Results: []model.SurveyResponse{
					{
						ID:            "resp-001",
						ParticipantID: "part-001",
						SubmissionID:  "sub-001",
						DateCreated:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			expectedContains: []string{"resp-001", "part-001", "sub-001"},
		},
		{
			name:           "GetSurveyResponses error",
			args:           []string{testSurveyID, "--table"},
			responsesError: errors.New("something went wrong"),
			expectedError:  "error: something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			c.EXPECT().
				GetSurveyResponses(gomock.Eq(testSurveyID), client.DefaultRecordLimit, client.DefaultRecordOffset).
				Return(tt.responsesReturn, tt.responsesError).
				Times(1)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := survey.NewResponseListCommand("list", c, writer)
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

			actual := b.String()

			if tt.expectedOutput != "" {
				if actual != tt.expectedOutput {
					t.Fatalf("expected\n'%s'\ngot\n'%s'\n", tt.expectedOutput, actual)
				}
			}

			for _, expected := range tt.expectedContains {
				if !bytes.Contains([]byte(actual), []byte(expected)) {
					t.Errorf("expected output to contain '%s', got:\n%s", expected, actual)
				}
			}
		})
	}
}
