package submission_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/submission"
	"github.com/prolific-oss/cli/mock_client"
)

const bulkApproveSuccessMessage = "The request to bulk approve has been made successfully.\n"

func TestNewBulkApproveCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := submission.NewBulkApproveCommand(c, os.Stdout)

	use := "bulk-approve"
	short := "Bulk approve multiple submissions"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestBulkApproveCommandErrorsIfNoFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		BulkApproveSubmissions(gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewBulkApproveCommand(c, writer)
	err := cmd.RunE(cmd, nil)

	writer.Flush()

	expected := "error: you must provide either --submission-id, --study and --participant-id, or --file"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestBulkApproveCommandErrorsIfMixedFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		BulkApproveSubmissions(gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewBulkApproveCommand(c, writer)
	_ = cmd.Flags().Set("submission-id", "sub-1")
	_ = cmd.Flags().Set("study", "study-1")
	err := cmd.RunE(cmd, nil)

	writer.Flush()

	expected := "error: cannot use --submission-id together with --study or --participant-id"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestBulkApproveCommandErrorsIfStudyWithoutParticipantIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		BulkApproveSubmissions(gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewBulkApproveCommand(c, writer)
	_ = cmd.Flags().Set("study", "study-1")
	err := cmd.RunE(cmd, nil)

	writer.Flush()

	expected := "error: --participant-id or --file is required when using --study"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestBulkApproveCommandErrorsIfParticipantIDsWithoutStudy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		BulkApproveSubmissions(gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewBulkApproveCommand(c, writer)
	_ = cmd.Flags().Set("participant-id", "part-1")
	err := cmd.RunE(cmd, nil)

	writer.Flush()

	expected := "error: --study is required when using --participant-id"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestBulkApproveCommandApprovesBySubmissionIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		BulkApproveSubmissions(gomock.Eq(client.BulkApproveSubmissionsPayload{
			SubmissionIDs: []string{"sub-1", "sub-2"},
		})).
		Return(nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewBulkApproveCommand(c, writer)
	_ = cmd.Flags().Set("submission-id", "sub-1")
	_ = cmd.Flags().Set("submission-id", "sub-2")
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("was not expected error, got %v", err)
	}

	writer.Flush()
	actual := b.String()

	if actual != bulkApproveSuccessMessage {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", bulkApproveSuccessMessage, actual)
	}
}

func TestBulkApproveCommandApprovesByStudyAndParticipantIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		BulkApproveSubmissions(gomock.Eq(client.BulkApproveSubmissionsPayload{
			StudyID:        "study-1",
			ParticipantIDs: []string{"part-1", "part-2"},
		})).
		Return(nil).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewBulkApproveCommand(c, writer)
	_ = cmd.Flags().Set("study", "study-1")
	_ = cmd.Flags().Set("participant-id", "part-1")
	_ = cmd.Flags().Set("participant-id", "part-2")
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("was not expected error, got %v", err)
	}

	writer.Flush()
	actual := b.String()

	if actual != bulkApproveSuccessMessage {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", bulkApproveSuccessMessage, actual)
	}
}

func TestBulkApproveCommandReturnsErrorIfApprovalFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	errorMessage := "Unable to bulk approve submissions"

	c.
		EXPECT().
		BulkApproveSubmissions(gomock.Eq(client.BulkApproveSubmissionsPayload{
			SubmissionIDs: []string{"sub-1"},
		})).
		Return(errors.New(errorMessage)).
		MaxTimes(1)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewBulkApproveCommand(c, writer)
	_ = cmd.Flags().Set("submission-id", "sub-1")
	err := cmd.RunE(cmd, nil)

	expected := fmt.Sprintf("error: %s", errorMessage)

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func TestBulkApproveCommandApprovesByFileAsSubmissionIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		BulkApproveSubmissions(gomock.Eq(client.BulkApproveSubmissionsPayload{
			SubmissionIDs: []string{"sub-1", "sub-2", "sub-3"},
		})).
		Return(nil).
		MaxTimes(1)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "submissions.csv")
	if err := os.WriteFile(filePath, []byte("sub-1\nsub-2\nsub-3\n"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewBulkApproveCommand(c, writer)
	_ = cmd.Flags().Set("file", filePath)
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("was not expected error, got %v", err)
	}

	writer.Flush()
	actual := b.String()

	if actual != bulkApproveSuccessMessage {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", bulkApproveSuccessMessage, actual)
	}
}

func TestBulkApproveCommandApprovesByFileAsParticipantIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		BulkApproveSubmissions(gomock.Eq(client.BulkApproveSubmissionsPayload{
			StudyID:        "study-1",
			ParticipantIDs: []string{"part-1", "part-2"},
		})).
		Return(nil).
		MaxTimes(1)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "participants.csv")
	if err := os.WriteFile(filePath, []byte("part-1\npart-2\n"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewBulkApproveCommand(c, writer)
	_ = cmd.Flags().Set("file", filePath)
	_ = cmd.Flags().Set("study", "study-1")
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("was not expected error, got %v", err)
	}

	writer.Flush()
	actual := b.String()

	if actual != bulkApproveSuccessMessage {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", bulkApproveSuccessMessage, actual)
	}
}

func TestBulkApproveCommandErrorsIfFileWithInlineIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		BulkApproveSubmissions(gomock.Any()).
		MaxTimes(0)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "submissions.csv")
	if err := os.WriteFile(filePath, []byte("sub-1\n"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewBulkApproveCommand(c, writer)
	_ = cmd.Flags().Set("file", filePath)
	_ = cmd.Flags().Set("submission-id", "sub-1")
	err := cmd.RunE(cmd, nil)

	writer.Flush()

	expected := "error: cannot use --file together with --submission-id or --participant-id"

	if err.Error() != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, err.Error())
	}
}

func assertBulkApproveFileError(t *testing.T, filePath, expectedSubstring string) {
	t.Helper()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		BulkApproveSubmissions(gomock.Any()).
		MaxTimes(0)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewBulkApproveCommand(c, writer)
	_ = cmd.Flags().Set("file", filePath)
	err := cmd.RunE(cmd, nil)

	writer.Flush()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), expectedSubstring) {
		t.Fatalf("expected error to contain '%s', got\n'%s'\n", expectedSubstring, err.Error())
	}
}

func TestBulkApproveCommandErrorsIfFileIsEmpty(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "empty.csv")
	if err := os.WriteFile(filePath, []byte(""), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	assertBulkApproveFileError(t, filePath, "file is empty")
}

func TestBulkApproveCommandErrorsIfFileNotFound(t *testing.T) {
	assertBulkApproveFileError(t, "/nonexistent/path/file.csv", "unable to read file")
}

func TestBulkApproveCommandFileSkipsBlankLines(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		BulkApproveSubmissions(gomock.Eq(client.BulkApproveSubmissionsPayload{
			SubmissionIDs: []string{"sub-1", "sub-2"},
		})).
		Return(nil).
		MaxTimes(1)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "submissions.csv")
	if err := os.WriteFile(filePath, []byte("sub-1\n\n  \nsub-2\n"), 0600); err != nil {
		t.Fatalf("failed to create temp file: %s", err)
	}

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewBulkApproveCommand(c, writer)
	_ = cmd.Flags().Set("file", filePath)
	err := cmd.RunE(cmd, nil)
	if err != nil {
		t.Fatalf("was not expected error, got %v", err)
	}
}
