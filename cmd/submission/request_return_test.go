package submission_test

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/prolific-oss/cli/client"
	"github.com/prolific-oss/cli/cmd/submission"
	"github.com/prolific-oss/cli/mock_client"
)

const testSubmissionID = "sub-123"

func TestNewRequestReturnCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	cmd := submission.NewRequestReturnCommand(c, os.Stdout)

	use := "request-return"
	short := "Request a participant to return a submission"

	if cmd.Use != use {
		t.Fatalf("expected use: %s; got %s", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Fatalf("expected short: %s; got %s", short, cmd.Short)
	}
}

func TestRequestReturnCommandCallsAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	returnTime := "2026-03-11T10:00:00Z"

	response := client.RequestSubmissionReturnResponse{
		ID:              testSubmissionID,
		Status:          "ACTIVE",
		Participant:     "participant-456",
		ReturnRequested: &returnTime,
	}

	c.
		EXPECT().
		RequestSubmissionReturn(
			gomock.Eq(testSubmissionID),
			gomock.Eq([]string{"Didn't finish the study"}),
		).
		Return(&response, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewRequestReturnCommand(c, writer)
	_ = cmd.Flags().Set("reason", "Didn't finish the study")
	_ = cmd.RunE(cmd, []string{testSubmissionID})
	writer.Flush()

	expected := "ID      Status Participant     Return Requested\nsub-123 ACTIVE participant-456 2026-03-11T10:00:00Z\n"
	actual := b.String()

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'", expected, actual)
	}
}

func TestRequestReturnCommandHandlesAPIError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	c.
		EXPECT().
		RequestSubmissionReturn(
			gomock.Eq(testSubmissionID),
			gomock.Eq([]string{"Withdrew consent"}),
		).
		Return(nil, errors.New("submission not found"))

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewRequestReturnCommand(c, writer)
	_ = cmd.Flags().Set("reason", "Withdrew consent")
	err := cmd.RunE(cmd, []string{testSubmissionID})
	writer.Flush()

	expected := "error: submission not found"
	if err == nil || err.Error() != expected {
		t.Fatalf("expected error '%s', got '%v'", expected, err)
	}
}

func TestRequestReturnCommandRequiresReason(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewRequestReturnCommand(c, writer)
	cmd.SetArgs([]string{testSubmissionID})
	err := cmd.Execute()

	writer.Flush()

	if err == nil {
		t.Fatal("expected error when reason flag is missing")
	}
}

func TestRequestReturnCommandWithNilReturnRequested(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	response := client.RequestSubmissionReturnResponse{
		ID:              testSubmissionID,
		Status:          "ACTIVE",
		Participant:     "participant-456",
		ReturnRequested: nil,
	}

	c.
		EXPECT().
		RequestSubmissionReturn(
			gomock.Eq(testSubmissionID),
			gomock.Eq([]string{"Didn't finish the study"}),
		).
		Return(&response, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewRequestReturnCommand(c, writer)
	_ = cmd.Flags().Set("reason", "Didn't finish the study")
	_ = cmd.RunE(cmd, []string{testSubmissionID})
	writer.Flush()

	expected := "ID      Status Participant     Return Requested\nsub-123 ACTIVE participant-456 -\n"
	actual := b.String()

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'", expected, actual)
	}
}

func TestRequestReturnCommandWithMultipleReasons(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := mock_client.NewMockAPI(ctrl)

	submissionID := "sub-789"
	returnTime := "2026-03-11T12:00:00Z"

	response := client.RequestSubmissionReturnResponse{
		ID:              submissionID,
		Status:          "ACTIVE",
		Participant:     "participant-101",
		ReturnRequested: &returnTime,
	}

	c.
		EXPECT().
		RequestSubmissionReturn(
			gomock.Eq(submissionID),
			gomock.Eq([]string{"Didn't finish the study", "Encountered technical problems"}),
		).
		Return(&response, nil)

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)

	cmd := submission.NewRequestReturnCommand(c, writer)
	_ = cmd.Flags().Set("reason", "Didn't finish the study")
	_ = cmd.Flags().Set("reason", "Encountered technical problems")
	_ = cmd.RunE(cmd, []string{submissionID})
	writer.Flush()

	expected := "ID      Status Participant     Return Requested\nsub-789 ACTIVE participant-101 2026-03-11T12:00:00Z\n"
	actual := b.String()

	if actual != expected {
		t.Fatalf("expected\n'%s'\ngot\n'%s'", expected, actual)
	}
}
