package submission_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/submission"
	"github.com/prolific-oss/cli/mock_client"
)

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

	expected := "error: you must provide either --submission-id or --study and --participant-id"

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

	expected := "error: --participant-id is required when using --study"

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

	expected := "The request to bulk approve has been made successfully.\n"

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
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

	expected := "The request to bulk approve has been made successfully.\n"

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'\n", expected, actual)
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
