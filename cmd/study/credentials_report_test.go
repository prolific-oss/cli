package study_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/study"
	"github.com/prolific-oss/cli/mock_client"
)

func TestNewCredentialsReportCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := study.NewCredentialsReportCommand(client, os.Stdout)

	use := "credentials-report"
	short := "Get the credentials usage report for a study"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestCredentialsReportCommand(t *testing.T) {
	studyID := "64395e9c2332b8a59a65d51e"

	tests := []struct {
		name           string
		studyID        string
		mockReturn     string
		mockError      error
		expectedOutput string
		expectedError  string
	}{
		{
			name:    "successful report retrieval",
			studyID: studyID,
			mockReturn: `Participant Id,Submission Id,Username,Status
68f8d0ca8efecd55f465ef02,68f8eb508f56127117b58491,fake_user_1,USED
68f8e1d75c5b0801988d377a,68f8f6ac43473bf358b2def1,fake_user_2,USED
,,fake_user_3,UNUSED
,,fake_user_4,UNUSED`,
			mockError: nil,
			expectedOutput: `Participant Id,Submission Id,Username,Status
68f8d0ca8efecd55f465ef02,68f8eb508f56127117b58491,fake_user_1,USED
68f8e1d75c5b0801988d377a,68f8f6ac43473bf358b2def1,fake_user_2,USED
,,fake_user_3,UNUSED
,,fake_user_4,UNUSED`,
			expectedError: "",
		},
		{
			name:           "study without credentials configured",
			studyID:        studyID,
			mockReturn:     "",
			mockError:      errors.New("request failed: Study does not have credentials configured"),
			expectedOutput: "",
			expectedError:  "request failed: Study does not have credentials configured",
		},
		{
			name:           "forbidden access",
			studyID:        studyID,
			mockReturn:     "",
			mockError:      errors.New("request failed with status 403: Forbidden"),
			expectedOutput: "",
			expectedError:  "request failed with status 403: Forbidden",
		},
		{
			name:           "study not found",
			studyID:        "invalid-id",
			mockReturn:     "",
			mockError:      errors.New("request failed with status 404: Not Found"),
			expectedOutput: "",
			expectedError:  "request failed with status 404: Not Found",
		},
		{
			name:           "empty report",
			studyID:        studyID,
			mockReturn:     `Participant Id,Submission Id,Username,Status`,
			mockError:      nil,
			expectedOutput: `Participant Id,Submission Id,Username,Status`,
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			c.
				EXPECT().
				GetStudyCredentialsUsageReportCSV(gomock.Eq(tt.studyID)).
				Return(tt.mockReturn, tt.mockError).
				Times(1)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := study.NewCredentialsReportCommand(c, writer)
			err := cmd.RunE(cmd, []string{tt.studyID})
			writer.Flush()

			// Check error expectation
			if tt.expectedError != "" {
				if err == nil {
					t.Fatalf("expected error '%s', got nil", tt.expectedError)
				}
				if err.Error() != tt.expectedError {
					t.Fatalf("expected error '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			// Check output expectation
			actual := b.String()
			if actual != tt.expectedOutput {
				t.Fatalf("expected output:\n'%s'\n\ngot:\n'%s'", tt.expectedOutput, actual)
			}
		})
	}
}
