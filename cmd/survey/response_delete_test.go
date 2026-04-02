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
)

func TestNewResponseDeleteCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := survey.NewResponseDeleteCommand("delete", c, os.Stdout)

	if cmd.Use != "delete <survey_id> <response_id>" {
		t.Fatalf("expected use: delete <survey_id> <response_id>; got %s", cmd.Use)
	}

	if cmd.Short != "Delete a survey response" {
		t.Fatalf("expected short: Delete a survey response; got %s", cmd.Short)
	}
}

func TestDeleteSurveyResponse(t *testing.T) {
	tests := []struct {
		name           string
		responseID     string
		mockError      error
		expectedOutput string
		expectedError  string
	}{
		{
			name:           "successful deletion",
			responseID:     "resp-001",
			mockError:      nil,
			expectedOutput: "Deleted survey response: resp-001\n",
		},
		{
			name:          "API error",
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
				DeleteSurveyResponse(gomock.Eq(testSurveyID), gomock.Eq(tt.responseID)).
				Return(tt.mockError).
				Times(1)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := survey.NewResponseDeleteCommand("delete", c, writer)
			cmd.SetArgs([]string{testSurveyID, tt.responseID})
			err := cmd.Execute()

			writer.Flush()

			if tt.expectedError != "" {
				if err == nil || err.Error() != tt.expectedError {
					t.Fatalf("expected error '%s'; got '%v'", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				actual := b.String()
				if actual != tt.expectedOutput {
					t.Fatalf("expected\n'%s'\ngot\n'%s'\n", tt.expectedOutput, actual)
				}
			}
		})
	}
}
