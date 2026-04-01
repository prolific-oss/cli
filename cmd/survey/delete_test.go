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

func TestNewDeleteCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := survey.NewDeleteCommand("delete", c, os.Stdout)

	if cmd.Use != "delete" {
		t.Fatalf("expected use: delete; got %s", cmd.Use)
	}

	if cmd.Short != "Delete a survey" {
		t.Fatalf("expected short: Delete a survey; got %s", cmd.Short)
	}
}

func TestDeleteSurvey(t *testing.T) {
	tests := []struct {
		name           string
		surveyID       string
		mockError      error
		expectedOutput string
		expectedError  string
	}{
		{
			name:           "successful deletion",
			surveyID:       "9eca0143-2gd0-44g4-b3e7-33f37ig763fb",
			mockError:      nil,
			expectedOutput: "Deleted survey: 9eca0143-2gd0-44g4-b3e7-33f37ig763fb\n",
		},
		{
			name:          "API error",
			surveyID:      "afdb1254-3he1-55h5-c4f8-44g48jh874gc",
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
				DeleteSurvey(gomock.Eq(tt.surveyID)).
				Return(tt.mockError).
				Times(1)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := survey.NewDeleteCommand("delete", c, writer)
			cmd.SetArgs([]string{tt.surveyID})
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
