package study_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/cmd/study"
	"github.com/prolific-oss/cli/mock_client"
	"github.com/prolific-oss/cli/model"
)

func TestNewSubmissionCountsCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	client := mock_client.NewMockAPI(ctrl)

	cmd := study.NewSubmissionCountsCommand(client, os.Stdout)

	use := "submission-counts"
	short := "Get submission counts by status for a study"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestSubmissionCountsNonInteractive(t *testing.T) {
	studyID := "64395e9c2332b8a59a65d51e"

	tests := []struct {
		name          string
		studyID       string
		mockReturn    *model.SubmissionCounts
		mockError     error
		expectOutput  bool
		expectedError string
	}{
		{
			name:    "successful counts retrieval",
			studyID: studyID,
			mockReturn: &model.SubmissionCounts{
				Active:            5,
				Approved:          10,
				AwaitingReview:    3,
				Rejected:          2,
				Reserved:          1,
				Returned:          0,
				TimedOut:          0,
				PartiallyApproved: 0,
				ScreenedOut:       0,
				Total:             21,
			},
			mockError:     nil,
			expectOutput:  true,
			expectedError: "",
		},
		{
			name:          "study not found",
			studyID:       "invalid-id",
			mockReturn:    nil,
			mockError:     errors.New("unable to fulfil request /api/v1/studies/invalid-id/submissions/counts/: not found"),
			expectOutput:  false,
			expectedError: "unable to fulfil request /api/v1/studies/invalid-id/submissions/counts/: not found",
		},
		{
			name:    "all zeros",
			studyID: studyID,
			mockReturn: &model.SubmissionCounts{
				Total: 0,
			},
			mockError:     nil,
			expectOutput:  true,
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			c := mock_client.NewMockAPI(ctrl)

			c.
				EXPECT().
				GetStudySubmissionCounts(gomock.Eq(tt.studyID)).
				Return(tt.mockReturn, tt.mockError).
				Times(1)

			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			cmd := study.NewSubmissionCountsCommand(c, writer)
			cmd.SetArgs([]string{tt.studyID, "-n"})
			err := cmd.Execute()
			writer.Flush()

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

			if tt.expectOutput {
				actual := b.String()
				if actual == "" {
					t.Fatal("expected output but got empty string")
				}
				if !bytes.Contains([]byte(actual), []byte("STATUS")) {
					t.Fatalf("expected output to contain STATUS header, got: %s", actual)
				}
				if !bytes.Contains([]byte(actual), []byte("Total")) {
					t.Fatalf("expected output to contain Total row, got: %s", actual)
				}
			}
		})
	}
}

func TestSubmissionCountsRendersCorrectValues(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

	c.
		EXPECT().
		GetStudySubmissionCounts(gomock.Eq(studyID)).
		Return(&model.SubmissionCounts{
			Active:            5,
			Approved:          10,
			AwaitingReview:    3,
			Rejected:          2,
			Reserved:          1,
			Returned:          0,
			TimedOut:          0,
			PartiallyApproved: 0,
			ScreenedOut:       0,
			Total:             21,
		}, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewSubmissionCountsCommand(c, writer)
	cmd.SetArgs([]string{studyID, "-n"})
	_ = cmd.Execute()
	writer.Flush()

	actual := b.String()

	expectedLines := []string{
		"STATUS               COUNT",
		"Active               5",
		"Approved             10",
		"Awaiting Review      3",
		"Rejected             2",
		"Reserved             1",
		"Returned             0",
		"Timed Out            0",
		"Partially Approved   0",
		"Screened Out         0",
		"Total                21",
	}

	for _, line := range expectedLines {
		if !strings.Contains(actual, line) {
			t.Fatalf("expected output to contain line '%s', got:\n%s", line, actual)
		}
	}
}

func TestSubmissionCountsJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	studyID := "11223344"

	expected := &model.SubmissionCounts{
		Active:            5,
		Approved:          10,
		AwaitingReview:    3,
		Rejected:          2,
		Reserved:          1,
		Returned:          0,
		TimedOut:          0,
		PartiallyApproved: 0,
		ScreenedOut:       0,
		Total:             21,
	}

	c.
		EXPECT().
		GetStudySubmissionCounts(gomock.Eq(studyID)).
		Return(expected, nil).
		Times(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := study.NewSubmissionCountsCommand(c, writer)
	cmd.SetArgs([]string{studyID, "--json"})
	err := cmd.Execute()
	writer.Flush()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var actual model.SubmissionCounts
	if err := json.Unmarshal(b.Bytes(), &actual); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput was: %s", err, b.String())
	}

	if actual != *expected {
		t.Fatalf("expected %+v, got %+v", *expected, actual)
	}
}
