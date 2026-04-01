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

func TestNewListCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := survey.NewListCommand("list", c, os.Stdout)

	if cmd.Use != "list" {
		t.Fatalf("expected use: list; got %s", cmd.Use)
	}

	if cmd.Short != "Provide a list of your surveys" {
		t.Fatalf("expected short: Provide a list of your surveys; got %s", cmd.Short)
	}
}

func TestListSurveys(t *testing.T) {
	tests := []struct {
		name           string
		meReturn       *client.MeResponse
		meError        error
		surveysReturn  *client.ListSurveysResponse
		surveysError   error
		expectedOutput string
		expectedError  string
	}{
		{
			name:     "successful list",
			meReturn: &client.MeResponse{ID: testResearcherID},
			surveysReturn: &client.ListSurveysResponse{
				Results: []model.Survey{
					{
						ID:          "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
						Title:       "Screening Survey",
						DateCreated: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:          "7ca8c921-0ebe-22e2-91c5-11d15ge541d9",
						Title:       "Follow-up Survey",
						DateCreated: time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			expectedOutput: `ID                                   Title            Date Created
6ba7b810-9dad-11d1-80b4-00c04fd430c8 Screening Survey 2026-01-15
7ca8c921-0ebe-22e2-91c5-11d15ge541d9 Follow-up Survey 2026-02-20

Showing 2 records of 2
`,
		},
		{
			name:          "GetMe error",
			meError:       errors.New("authentication failed"),
			expectedError: "error: authentication failed",
		},
		{
			name:          "GetSurveys error",
			meReturn:      &client.MeResponse{ID: testResearcherID},
			surveysError:  errors.New("something went wrong"),
			expectedError: "error: something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			c.EXPECT().
				GetMe().
				Return(tt.meReturn, tt.meError).
				Times(1)

			if tt.meError == nil {
				c.EXPECT().
					GetSurveys(gomock.Eq(testResearcherID), client.DefaultRecordLimit, client.DefaultRecordOffset).
					Return(tt.surveysReturn, tt.surveysError).
					Times(1)
			}

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := survey.NewListCommand("list", c, writer)
			err := cmd.RunE(cmd, nil)

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
			if actual != tt.expectedOutput {
				t.Fatalf("expected\n'%s'\ngot\n'%s'\n", tt.expectedOutput, actual)
			}
		})
	}
}
